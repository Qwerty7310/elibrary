package postgres

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/readmodel"
	"elibrary/internal/repository"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookRepository struct {
	db *pgxpool.Pool
}

func NewBookRepository(db *pgxpool.Pool) *BookRepository {
	return &BookRepository{db: db}
}

func (r *BookRepository) Create(ctx context.Context, book domain.Book) error {
	extraJSON, err := json.Marshal(book.Extra)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO books (id, barcode, factory_barcode, title, publisher_id, year, description, location_id, extra)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		book.ID,
		book.Barcode,
		book.FactoryBarcode,
		book.Title,
		book.PublisherID,
		book.Year,
		book.Description,
		book.LocationID,
		extraJSON,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrBarcodeExists
		}
		return err
	}

	return nil
}

func (r *BookRepository) Update(ctx context.Context, book domain.Book) error {
	extraJSON, err := json.Marshal(book.Extra)
	if err != nil {
		return err
	}

	res, err := r.db.Exec(ctx, `
		UPDATE books
		SET
		    barcode = $2,
		    factory_barcode = $3,
		    title = $4,
		    publisher_id = $5,
		    year = $6,
		    description = $7,
		    location_id = $8,
		    extra = $9,
		    updated_at = NOW()
		WHERE id = $1
	`,
		book.ID,
		book.Barcode,
		book.FactoryBarcode,
		book.Title,
		book.PublisherID,
		book.Year,
		book.Description,
		book.LocationID,
		extraJSON,
	)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *BookRepository) GetPublicByID(ctx context.Context, id uuid.UUID) (*readmodel.BookPublic, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var book readmodel.BookPublic
	var extraJSON []byte

	err = loadBookBase(
		ctx, tx, id,
		&book.ID,
		&book.Barcode,
		&book.FactoryBarcode,
		&book.Title,
		&book.Year,
		&book.Description,
		&extraJSON,
		&book.CreatedAt,
		&book.UpdatedAt,
		&book.Publisher,
	)
	if err != nil {
		return nil, err
	}

	if len(extraJSON) > 0 {
		if err := json.Unmarshal(extraJSON, &book.Extra); err != nil {
			return nil, err
		}
	} else {
		book.Extra = make(map[string]any)
	}

	if err := loadWorks(ctx, tx, id, &book.Works); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &book, nil
}

func (r *BookRepository) GetInternalByID(ctx context.Context, id uuid.UUID) (*readmodel.BookInternal, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var book readmodel.BookInternal
	var extraJSON []byte

	err = loadBookBase(
		ctx, tx, id,
		&book.ID,
		&book.Barcode,
		&book.FactoryBarcode,
		&book.Title,
		&book.Year,
		&book.Description,
		&extraJSON,
		&book.CreatedAt,
		&book.UpdatedAt,
		&book.Publisher,
	)
	if err != nil {
		return nil, err
	}

	if len(extraJSON) > 0 {
		if err := json.Unmarshal(extraJSON, &book.Extra); err != nil {
			return nil, err
		}
	} else {
		book.Extra = make(map[string]any)
	}

	if err := loadWorks(ctx, tx, id, &book.Works); err != nil {
		return nil, err
	}

	if err := loadLocation(ctx, tx, id, &book.Location); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &book, nil
}

func loadBookBase(
	ctx context.Context,
	tx pgx.Tx,
	id uuid.UUID,
	bookID *uuid.UUID,
	barcode *string,
	factoryBarcode **string,
	title *string,
	year **int,
	description **string,
	extraJSON *[]byte,
	createdAt *time.Time,
	updatedAt *time.Time,
	publisher **readmodel.Publisher,
) error {
	var (
		publisherID   *uuid.UUID
		publisherName *string
	)

	err := tx.QueryRow(ctx, `
		SELECT
			b.id,
			b.barcode,
			b.factory_barcode,
			b.title,
		    b.year,
		    b.description,
		    b.extra,
		    b.created_at,
		    b.updated_at,
		    p.id,
		    p.name
		FROM books b
		LEFT JOIN publishers p ON p.id = b.publisher_id
		WHERE b.id = $1
	`, id).Scan(
		bookID,
		barcode,
		factoryBarcode,
		title,
		year,
		description,
		extraJSON,
		createdAt,
		updatedAt,
		&publisherID,
		&publisherName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrNotFound
		}
		return err
	}

	if publisherID != nil && publisherName != nil {
		*publisher = &readmodel.Publisher{
			ID:   *publisherID,
			Name: *publisherName,
		}
	}

	return nil
}

func loadWorks(
	ctx context.Context,
	tx pgx.Tx,
	bookID uuid.UUID,
	target *[]*readmodel.WorkShort,
) error {
	rows, err := tx.Query(ctx, `
		SELECT
			w.id,
			w.title,
			a.id,
			a.last_name,
			a.first_name,
			a.middle_name
		FROM book_works bw
		JOIN works w ON w.id = bw.work_id
		LEFT JOIN work_authors wa ON wa.work_id = w.id
		LEFT JOIN authors a ON a.id = wa.author_id
		WHERE bw.book_id = $1
		ORDER BY bw.position NULLS LAST, w.title
	`, bookID)
	if err != nil {
		return err
	}
	defer rows.Close()

	workMap := make(map[uuid.UUID]*readmodel.WorkShort)
	for rows.Next() {
		var (
			workID uuid.UUID
			title  string

			authorID   *uuid.UUID
			lastName   *string
			firstName  *string
			middleName *string
		)

		if err := rows.Scan(
			&workID,
			&title,
			&authorID,
			&lastName,
			&firstName,
			&middleName,
		); err != nil {
			return err
		}

		work, ok := workMap[workID]
		if !ok {
			work = &readmodel.WorkShort{
				ID:    workID,
				Title: title,
			}
			workMap[workID] = work
			*target = append(*target, work)

		}

		if authorID != nil {
			work.Authors = append(work.Authors, readmodel.Author{
				ID:         *authorID,
				LastName:   derefStr(lastName),
				FirstName:  firstName,
				MiddleName: middleName,
			})
		}
	}

	return rows.Err()
}

func loadLocation(
	ctx context.Context,
	tx pgx.Tx,
	bookID uuid.UUID,
	target **readmodel.Location,
) error {
	var (
		shelfID   *uuid.UUID
		shelfName *string

		cabinetID   *uuid.UUID
		cabinetName *string

		roomID   *uuid.UUID
		roomName *string

		buildingID   *uuid.UUID
		buildingName *string
		address      *string
	)

	err := tx.QueryRow(ctx, `
		SELECT
		    s.id, s.name,
		    c.id, c.name,
		    r.id, r.name,
		    b.id, b.name, b.address
		FROM books bk
		LEFT JOIN locations s ON s.id = bk.location_id AND s.type = 'shelf'
		LEFT JOIN locations c ON c.id = s.parent_id AND c.type = 'cabinet'
		LEFT JOIN locations r ON r.id = c.parent_id AND r.type = 'room'
		LEFT JOIN locations b ON b.id = r.parent_id AND b.type = 'building'
		WHERE bk.id = $1
	`, bookID).Scan(
		&shelfID, &shelfName,
		&cabinetID, &cabinetName,
		&roomID, &roomName,
		&buildingID, &buildingName,
		&address,
	)
	if err != nil {
		return err
	}

	if shelfID != nil {
		*target = &readmodel.Location{
			ShelfID:   derefUUID(shelfID),
			ShelfName: derefStr(shelfName),

			CabinetID:   derefUUID(cabinetID),
			CabinetName: derefStr(cabinetName),

			RoomID:   derefUUID(roomID),
			RoomName: derefStr(roomName),

			BuildingID:   derefUUID(buildingID),
			BuildingName: derefStr(buildingName),

			Address: derefStr(address),
		}
	}

	return nil
}

func derefStr(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func derefUUID(id *uuid.UUID) uuid.UUID {
	if id != nil {
		return *id
	}
	return uuid.Nil
}

func (r *BookRepository) GetPublic(ctx context.Context, filter repository.BookFilter) ([]*readmodel.BookPublic, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	base, err := r.getBooksBase(ctx, tx, filter)
	if err != nil {
		return nil, err
	}

	var res []*readmodel.BookPublic
	for _, book := range base {
		res = append(res, &readmodel.BookPublic{
			ID:             book.ID,
			Title:          book.Title,
			Barcode:        book.Barcode,
			FactoryBarcode: book.FactoryBarcode,
			Publisher:      book.Publisher,
			Works:          book.Works,
			Year:           book.Year,
			Description:    book.Description,
			Extra:          book.Extra,
			CreatedAt:      book.CreatedAt,
			UpdatedAt:      book.UpdatedAt,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *BookRepository) GetInternal(ctx context.Context, filter repository.BookFilter) ([]*readmodel.BookInternal, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	base, err := r.getBooksBase(ctx, tx, filter)
	if err != nil {
		return nil, err
	}

	if err := loadLocationsForBooks(ctx, tx, base); err != nil {
		return nil, err
	}

	res := make([]*readmodel.BookInternal, 0, len(base))
	for _, book := range base {
		res = append(res, &readmodel.BookInternal{
			ID:             book.ID,
			Title:          book.Title,
			Barcode:        book.Barcode,
			FactoryBarcode: book.FactoryBarcode,
			Publisher:      book.Publisher,
			Location:       book.Location,
			Works:          book.Works,
			Year:           book.Year,
			Description:    book.Description,
			Extra:          book.Extra,
			CreatedAt:      book.CreatedAt,
			UpdatedAt:      book.UpdatedAt,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *BookRepository) getBooksBase(
	ctx context.Context,
	tx pgx.Tx,
	filter repository.BookFilter,
) ([]*bookBase, error) {

	rows, err := tx.Query(ctx, `
		SELECT
			b.id,
			b.barcode,
			b.factory_barcode,
			b.title,
			b.year,
			b.description,
			b.extra,
			b.created_at,
			b.updated_at,
			p.id,
			p.name
		FROM books b
		LEFT JOIN publishers p ON p.id = b.publisher_id
		WHERE
		    (
		        ($1::uuid IS NOT NULL AND b.id = $1)
		        OR
		        ($1::uuid IS NULL AND $2::text IS NOT NULL AND b.barcode = $2)
		        OR
		        ($1::uuid IS NULL AND $2::text IS NULL AND $3::text IS NOT NULL AND b.factory_barcode = $3)
		        OR
		        (
		            $1::uuid IS NULL
					AND $2::text IS NULL
		            AND $3::text IS NULL
		            AND ($4::text IS NULL OR b.search_vector @@ plainto_tsquery('russian', $4))
		        )
		    )
			AND ($5::uuid IS NULL OR b.publisher_id = $5)
			AND ($6::int IS NULL OR b.year >= $6)
			AND ($7::int IS NULL OR b.year <= $7)
		ORDER BY b.created_at DESC
		LIMIT $8 OFFSET $9
	`,
		filter.ID,
		filter.Barcode,
		filter.FactoryBarcode,
		filter.Query,
		filter.PublisherID,
		filter.YearFrom,
		filter.YearTo,
		filter.LimitOr(20),
		filter.OffsetOr(0),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []*bookBase

	for rows.Next() {
		var (
			book          bookBase
			extraJSON     []byte
			publisherID   *uuid.UUID
			publisherName *string
		)

		if err := rows.Scan(
			&book.ID,
			&book.Barcode,
			&book.FactoryBarcode,
			&book.Title,
			&book.Year,
			&book.Description,
			&extraJSON,
			&book.CreatedAt,
			&book.UpdatedAt,
			&publisherID,
			&publisherName,
		); err != nil {
			return nil, err
		}

		if publisherID != nil {
			book.Publisher = &readmodel.Publisher{
				ID:   *publisherID,
				Name: derefStr(publisherName),
			}
		}

		if len(extraJSON) > 0 {
			if err := json.Unmarshal(extraJSON, &book.Extra); err != nil {
				return nil, err
			}
		} else {
			book.Extra = make(map[string]any)
		}

		books = append(books, &book)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(books) == 0 {
		return nil, repository.ErrNotFound
	}

	if err := loadWorksForBooks(ctx, tx, books); err != nil {
		return nil, err
	}

	return books, nil
}

func loadWorksForBooks(ctx context.Context, tx pgx.Tx, books []*bookBase) error {
	if len(books) == 0 {
		return nil
	}

	bookMap := make(map[uuid.UUID]*bookBase, len(books))
	bookIDs := make([]uuid.UUID, 0, len(books))

	for _, book := range books {
		bookMap[book.ID] = book
		bookIDs = append(bookIDs, book.ID)
	}

	rows, err := tx.Query(ctx, `
		SELECT
		    bw.book_id,
		    w.id,
		    w.title,
		    a.id,
		    a.last_name,
		    a.first_name,
		    a.middle_name
		FROM book_works bw
		JOIN works w ON w.id = bw.work_id
		LEFT JOIN work_authors wa ON wa.work_id = w.id
		LEFT JOIN authors a ON a.id = wa.author_id
		WHERE bw.book_id = ANY($1)
		ORDER BY
		    bw.book_id,
		    bw.position NULLS LAST,
		    w.title
	`, bookIDs)
	if err != nil {
		return err
	}
	defer rows.Close()

	type workKey struct {
		bookID uuid.UUID
		workID uuid.UUID
	}

	workMap := make(map[workKey]*readmodel.WorkShort)

	for rows.Next() {
		var (
			bookID uuid.UUID
			workID uuid.UUID
			title  string

			authorID   *uuid.UUID
			lastName   *string
			firstName  *string
			middleName *string
		)

		if err := rows.Scan(
			&bookID,
			&workID,
			&title,
			&authorID,
			&lastName,
			&firstName,
			&middleName,
		); err != nil {
			return err
		}

		book, ok := bookMap[bookID]
		if !ok {
			continue
		}

		key := workKey{bookID, workID}
		work, ok := workMap[key]
		if !ok {
			work = &readmodel.WorkShort{
				ID:    workID,
				Title: title,
			}
			workMap[key] = work
			book.Works = append(book.Works, work)
		}

		if authorID != nil {
			work.Authors = append(work.Authors, readmodel.Author{
				ID:         *authorID,
				LastName:   derefStr(lastName),
				FirstName:  firstName,
				MiddleName: middleName,
			})
		}
	}

	return rows.Err()
}

func loadLocationsForBooks(ctx context.Context, tx pgx.Tx, books []*bookBase) error {
	if len(books) == 0 {
		return nil
	}
	bookMap := make(map[uuid.UUID]*bookBase, len(books))
	bookIDs := make([]uuid.UUID, 0, len(books))

	for _, book := range books {
		bookMap[book.ID] = book
		bookIDs = append(bookIDs, book.ID)
	}

	rows, err := tx.Query(ctx, `
		SELECT
		    bk.id,
		    
		    s.id, s.name,
		    c.id, c.name,
		    r.id, r.name,
		    b.id, b.name,
		    b.address
		FROM books bk
		LEFT JOIN locations s ON s.id = bk.location_id AND s.type = 'shelf'
		LEFT JOIN locations c ON c.id = s.parent_id AND c.type = 'cabinet'
		LEFT JOIN locations r ON r.id = c.parent_id AND r.type = 'room'
		LEFT JOIN locations b ON b.id = r.parent_id AND b.type = 'building'
		WHERE bk.id = ANY($1)
	`, bookIDs)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			bookID uuid.UUID

			shelfID   *uuid.UUID
			shelfName *string

			cabinetID   *uuid.UUID
			cabinetName *string

			roomID   *uuid.UUID
			roomName *string

			buildingID   *uuid.UUID
			buildingName *string
			address      *string
		)

		if err := rows.Scan(
			&bookID,

			&shelfID, &shelfName,
			&cabinetID, &cabinetName,
			&roomID, &roomName,
			&buildingID, &buildingName,
			&address,
		); err != nil {
			return err
		}

		if shelfID == nil {
			continue
		}

		book := bookMap[bookID]
		book.Location = &readmodel.Location{
			ShelfID:   derefUUID(shelfID),
			ShelfName: derefStr(shelfName),

			CabinetID:   derefUUID(cabinetID),
			CabinetName: derefStr(cabinetName),

			RoomID:   derefUUID(roomID),
			RoomName: derefStr(roomName),

			BuildingID:   derefUUID(buildingID),
			BuildingName: derefStr(buildingName),
			Address:      derefStr(address),
		}
	}

	return rows.Err()
}

func (r *BookRepository) WithTx(ctx context.Context, fn func(tx repository.BookTx) error) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	wrapped := &bookTx{tx: tx}

	if err := fn(wrapped); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

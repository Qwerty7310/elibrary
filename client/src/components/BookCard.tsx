// src/components/BookCard.tsx
import type { Book } from "../types/book"

interface BookCardProps {
    book: Book
    onViewDetails?: (book: Book) => void
}

export function BookCard({ book, onViewDetails }: BookCardProps) {
    return (
        <div className="book-card">
            <div className="book-info">
                <h3>{book.title}</h3>
                <p>Автор: {book.author}</p>
                {book.publisher && <p>Издательство: {book.publisher}</p>}
                {book.year && <p>Год: {book.year}</p>}
                {book.location && <p>Местоположение: {book.location}</p>}
            </div>
            <div className="book-actions">
                <button onClick={() => onViewDetails?.(book)}>
                    Подробнее
                </button>
                <button onClick={() => window.open(`/books/${book.id}/barcode`, '_blank')}>
                    Штрихкод
                </button>
            </div>
        </div>
    )
}
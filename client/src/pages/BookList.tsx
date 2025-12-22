import {type FormEvent, useState} from "react"
import type {Book} from "../types/book"
import {scanBook, searchBooks} from "../api/books"
import {SearchBar} from "../components/SearchBar"

type DraftBook = {
    title: string
    author: string
    barcode: string
    publisher: string
    year: string
    location: string
}

const emptyDraft: DraftBook = {
    title: "",
    author: "",
    barcode: "",
    publisher: "",
    year: "",
    location: "",
}

function escapeRegExp(value: string) {
    return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")
}

function highlightText(text: string, query: string) {
    const trimmed = query.trim()
    if (!trimmed) {
        return text
    }

    const regex = new RegExp(`(${escapeRegExp(trimmed)})`, "ig")
    const parts = text.split(regex)

    return parts.map((part, index) =>
        index % 2 === 1 ? (
            <mark key={`${part}-${index}`} className="highlight">
                {part}
            </mark>
        ) : (
            <span key={`${part}-${index}`}>{part}</span>
        )
    )
}

function createLocalId() {
    if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
        return crypto.randomUUID()
    }
    return `local-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

export function BookList() {
    const [books, setBooks] = useState<Book[]>([])
    const [query, setQuery] = useState("")
    const [error, setError] = useState<string | null>(null)
    const [isLoading, setIsLoading] = useState(false)
    const [selectedBook, setSelectedBook] = useState<Book | null>(null)
    const [isAddOpen, setIsAddOpen] = useState(false)
    const [draft, setDraft] = useState<DraftBook>(emptyDraft)

    async function handleSearch(value: string) {
        const trimmed = value.trim()
        setError(null)
        setQuery(value)

        if (!trimmed) {
            setBooks([])
            return
        }

        setIsLoading(true)
        try {
            // сначала пробуем как скан
            const scannedBook = await scanBook(trimmed)
            setBooks([scannedBook])
            return
        } catch {
            try {
                // если не нашли — обычный поиск
                const result = await searchBooks(trimmed)
                setBooks(result)
                if (result.length === 0) {
                    setError("Ничего не найдено")
                }
            } catch {
                setError("Не удалось выполнить поиск")
            }
        } finally {
            setIsLoading(false)
        }
    }

    function handleOpenAdd() {
        setDraft(emptyDraft)
        setIsAddOpen(true)
    }

    function handleAddSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        const yearNumber = draft.year ? Number(draft.year) : undefined
        const newBook: Book = {
            id: createLocalId(),
            barcode: draft.barcode,
            title: draft.title,
            author: draft.author,
            publisher: draft.publisher || undefined,
            year: Number.isFinite(yearNumber) ? yearNumber : undefined,
            location: draft.location || undefined,
        }

        setBooks((prev) => [newBook, ...prev])
        setIsAddOpen(false)
        setSelectedBook(newBook)
        setError(null)
    }

    const resultsLabel =
        books.length === 0
            ? "Пока ничего не найдено"
            : `Найдено: ${books.length}`

    const detailsQuery = query.trim()

    function renderValue(value?: string | number) {
        if (value === undefined || value === null || value === "") {
            return "нет данных"
        }
        return typeof value === "string"
            ? highlightText(value, detailsQuery)
            : value
    }

    return (
        <div className="app-shell">
            <header className="hero">
                <div className="hero-copy">
                    <p className="eyebrow">Электронная библиотека</p>
                    <h1>Найдите книгу, которой хочется делиться</h1>
                    <p className="hero-subtitle">
                        Ищите по автору, названию или штрихкоду. Добавляйте новые
                        издания, чтобы коллекция всегда оставалась актуальной.
                    </p>
                </div>
                <div className="hero-actions">
                    <SearchBar onSearch={handleSearch} isLoading={isLoading} />
                    <button
                        className="primary-button"
                        type="button"
                        onClick={handleOpenAdd}
                    >
                        Добавить книгу
                    </button>
                </div>
            </header>

            <section className="results">
                <div className="results-header">
                    <div>
                        <h2>Результаты поиска</h2>
                        <p className="results-caption">{resultsLabel}</p>
                    </div>
                    {detailsQuery && (
                        <span className="query-chip">
                            Поиск: "{detailsQuery}"
                        </span>
                    )}
                </div>

                {error && <p className="error-banner">{error}</p>}

                {isLoading && <p className="status-line">Идет поиск…</p>}

                {!isLoading && books.length === 0 && !error && (
                    <div className="empty-state">
                        <p>Введите запрос и нажмите Enter, чтобы начать поиск.</p>
                    </div>
                )}

                <div className="book-grid">
                    {books.map((book) => (
                        <button
                            className="book-card"
                            type="button"
                            key={book.id}
                            onClick={() => setSelectedBook(book)}
                        >
                            <div className="book-meta">
                                <p className="book-title">
                                    {highlightText(book.title, detailsQuery)}
                                </p>
                                <p className="book-author">
                                    {highlightText(book.author, detailsQuery)}
                                </p>
                            </div>
                            <div className="book-tags">
                                <span>{book.publisher || "Без издателя"}</span>
                                <span>{book.year || "Год не указан"}</span>
                            </div>
                        </button>
                    ))}
                </div>
            </section>

            {selectedBook && (
                <div
                    className="modal-backdrop"
                    role="dialog"
                    aria-modal="true"
                    onClick={(event) => {
                        if (event.target === event.currentTarget) {
                            setSelectedBook(null)
                        }
                    }}
                >
                    <div className="modal-card">
                        <header className="modal-header">
                            <div>
                                <p className="modal-eyebrow">Карточка книги</p>
                                <h3>{highlightText(selectedBook.title, detailsQuery)}</h3>
                            </div>
                            <button
                                className="ghost-button"
                                type="button"
                                onClick={() => setSelectedBook(null)}
                            >
                                Закрыть
                            </button>
                        </header>
                        <div className="modal-body">
                            <div className="detail-row">
                                <span>Автор</span>
                                <strong>{renderValue(selectedBook.author)}</strong>
                            </div>
                            <div className="detail-row">
                                <span>Штрихкод</span>
                                <strong>{renderValue(selectedBook.barcode)}</strong>
                            </div>
                            <div className="detail-row">
                                <span>Издательство</span>
                                <strong>{renderValue(selectedBook.publisher)}</strong>
                            </div>
                            <div className="detail-row">
                                <span>Год</span>
                                <strong>{renderValue(selectedBook.year)}</strong>
                            </div>
                            <div className="detail-row">
                                <span>Место хранения</span>
                                <strong>{renderValue(selectedBook.location)}</strong>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {isAddOpen && (
                <div
                    className="modal-backdrop"
                    role="dialog"
                    aria-modal="true"
                    onClick={(event) => {
                        if (event.target === event.currentTarget) {
                            setIsAddOpen(false)
                        }
                    }}
                >
                    <div className="modal-card">
                        <header className="modal-header">
                            <div>
                                <p className="modal-eyebrow">Новая книга</p>
                                <h3>Добавить запись</h3>
                            </div>
                            <button
                                className="ghost-button"
                                type="button"
                                onClick={() => setIsAddOpen(false)}
                            >
                                Закрыть
                            </button>
                        </header>
                        <form className="modal-body" onSubmit={handleAddSubmit}>
                            <label className="field">
                                Название
                                <input
                                    value={draft.title}
                                    onChange={(event) =>
                                        setDraft((prev) => ({
                                            ...prev,
                                            title: event.target.value,
                                        }))
                                    }
                                    required
                                />
                            </label>
                            <label className="field">
                                Автор
                                <input
                                    value={draft.author}
                                    onChange={(event) =>
                                        setDraft((prev) => ({
                                            ...prev,
                                            author: event.target.value,
                                        }))
                                    }
                                    required
                                />
                            </label>
                            <label className="field">
                                Штрихкод
                                <input
                                    value={draft.barcode}
                                    onChange={(event) =>
                                        setDraft((prev) => ({
                                            ...prev,
                                            barcode: event.target.value,
                                        }))
                                    }
                                    required
                                />
                            </label>
                            <label className="field">
                                Издательство
                                <input
                                    value={draft.publisher}
                                    onChange={(event) =>
                                        setDraft((prev) => ({
                                            ...prev,
                                            publisher: event.target.value,
                                        }))
                                    }
                                />
                            </label>
                            <label className="field">
                                Год
                                <input
                                    type="number"
                                    inputMode="numeric"
                                    value={draft.year}
                                    onChange={(event) =>
                                        setDraft((prev) => ({
                                            ...prev,
                                            year: event.target.value,
                                        }))
                                    }
                                />
                            </label>
                            <label className="field">
                                Место хранения
                                <input
                                    value={draft.location}
                                    onChange={(event) =>
                                        setDraft((prev) => ({
                                            ...prev,
                                            location: event.target.value,
                                        }))
                                    }
                                />
                            </label>
                            <div className="modal-actions">
                                <button className="primary-button" type="submit">
                                    Сохранить
                                </button>
                                <button
                                    className="ghost-button"
                                    type="button"
                                    onClick={() => setIsAddOpen(false)}
                                >
                                    Отмена
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    )
}

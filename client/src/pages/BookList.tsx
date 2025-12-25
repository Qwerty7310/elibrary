import {type FormEvent, useCallback, useEffect, useRef, useState} from "react"
import type {Book} from "../types/book"
import {createBook, scanBooks, searchBooks} from "../api/books"
import {SearchBar} from "../components/SearchBar"

type ExtraField = {
    id: string
    key: string
    value: string
}

type DraftBook = {
    title: string
    author: string
    factoryBarcode: string
    publisher: string
    year: string
    location: string
    extra: ExtraField[]
}

const emptyDraft: DraftBook = {
    title: "",
    author: "",
    factoryBarcode: "",
    publisher: "",
    year: "",
    location: "",
    extra: [{id: createExtraId(), key: "", value: ""}],
}

const MAX_QUERY_LENGTH = 100
const MAX_QUERY_LABEL = 48

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

function truncateLabel(value: string, maxLength: number) {
    if (value.length <= maxLength) {
        return value
    }
    return `${value.slice(0, Math.max(0, maxLength - 1))}…`
}

function isValidUuid(value: string) {
    return /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(
        value
    )
}

function isValidEAN13(value: string) {
    if (!/^\d{13}$/.test(value)) {
        return false
    }
    const digits = value.split("").map((digit) => Number(digit))
    const checksum = digits.pop()
    if (checksum === undefined) {
        return false
    }
    const sum = digits.reduce(
        (acc, digit, index) => acc + digit * (index % 2 === 0 ? 1 : 3),
        0
    )
    const calculated = (10 - (sum % 10)) % 10
    return calculated === checksum
}

function createExtraId() {
    if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
        return crypto.randomUUID()
    }
    return `extra-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

export function BookList() {
    const [books, setBooks] = useState<Book[]>([])
    const [query, setQuery] = useState("")
    const [error, setError] = useState<string | null>(null)
    const [isLoading, setIsLoading] = useState(false)
    const [selectedBook, setSelectedBook] = useState<Book | null>(null)
    const [isAddOpen, setIsAddOpen] = useState(false)
    const [draft, setDraft] = useState<DraftBook>(emptyDraft)
    const [addError, setAddError] = useState<string | null>(null)
    const [isAddSaving, setIsAddSaving] = useState(false)
    const [isScanOpen, setIsScanOpen] = useState(false)
    const [scanValue, setScanValue] = useState("")
    const [scanMessage, setScanMessage] = useState<string | null>(null)
    const [isScanLoading, setIsScanLoading] = useState(false)
    const [scanHighlight, setScanHighlight] = useState("")
    const [searchMode, setSearchMode] = useState<"text" | "barcode" | null>(
        null
    )
    const scanInputRef = useRef<HTMLInputElement | null>(null)

    async function handleSearch(value: string) {
        const trimmed = value.trim()
        setError(null)
        setScanHighlight("")
        setSearchMode("text")
        setQuery(value)

        if (!trimmed) {
            setBooks([])
            return
        }

        if (trimmed.length > MAX_QUERY_LENGTH) {
            setError(
                `Превышена допустимая длина запроса (${MAX_QUERY_LENGTH} символов).`
            )
            return
        }

        setIsLoading(true)
        try {
            const result = await searchBooks(trimmed)
            setBooks(result)
            if (result.length === 0) {
                setError("Ничего не найдено")
            }
        } catch {
            setError("Не удалось выполнить поиск")
        } finally {
            setIsLoading(false)
        }
    }

    function handleOpenAdd() {
        setDraft(emptyDraft)
        setAddError(null)
        setIsAddOpen(true)
    }

    async function handleAddSubmit() {
        setAddError(null)
        setIsAddSaving(true)
        const yearNumber = draft.year ? Number(draft.year) : undefined
        const extra = draft.extra.reduce<Record<string, unknown>>(
            (acc, field) => {
                const key = field.key.trim()
                if (!key) {
                    return acc
                }
                acc[key] = field.value.trim()
                return acc
            },
            {}
        )

        try {
            const created = await createBook({
                title: draft.title.trim(),
                author: draft.author.trim(),
                publisher: draft.publisher.trim() || undefined,
                year: Number.isFinite(yearNumber) ? yearNumber : undefined,
                location: draft.location.trim() || undefined,
                factory_barcode: draft.factoryBarcode.trim() || undefined,
                extra: Object.keys(extra).length > 0 ? extra : undefined,
            })

            setIsAddOpen(false)
            setSelectedBook(created)
            setError(null)
        } catch (err) {
            const status = (err as Error & {status?: number}).status
            if (status === 409) {
                setAddError("Штрихкод уже существует")
            } else if (status === 400) {
                setAddError("Некорректные данные для книги")
            } else {
                setAddError("Не удалось сохранить книгу")
            }
        } finally {
            setIsAddSaving(false)
        }
    }

    const submitScan = useCallback(
        async (value: string) => {
            const trimmed = value.trim()
            if (!trimmed) {
                setScanMessage("Введите штрихкод для поиска.")
                return
            }

            const valid = isValidEAN13(trimmed) || isValidUuid(trimmed)
            if (!valid) {
                setScanMessage("Штрихкод невалиден. Проверьте формат EAN-13.")
                return
            }

            setIsScanLoading(true)
            setScanMessage(null)
            try {
                const result = await scanBooks(trimmed)
                setBooks(result)
                setQuery("")
                setSearchMode("barcode")
                setScanHighlight(trimmed)
                setSelectedBook(null)
                setError(null)
                setIsScanOpen(false)
            } catch (err) {
                const status = (err as Error & {status?: number}).status
                if (status === 400) {
                    setScanMessage("Штрихкод невалиден. Проверьте формат EAN-13.")
                } else if (status === 404) {
                    setIsScanOpen(false)
                    setDraft((prev) => ({
                        ...prev,
                        factoryBarcode: trimmed,
                    }))
                    setIsAddOpen(true)
                } else {
                    setScanMessage("Не удалось выполнить поиск по штрихкоду.")
                }
            } finally {
                setIsScanLoading(false)
            }
        },
        [setBooks, setError]
    )

    function handleScanSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        void submitScan(scanValue)
    }

    function handleOpenScan() {
        setScanValue("")
        setScanMessage(null)
        setIsScanOpen(true)
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

    function renderBarcodeValue(value?: string) {
        if (!value) {
            return "нет данных"
        }
        return highlightText(value, scanHighlight || detailsQuery)
    }

    function renderExtraValue(value: unknown) {
        if (value === undefined || value === null || value === "") {
            return "нет данных"
        }
        if (typeof value === "string") {
            return value
        }
        return String(value)
    }

    useEffect(() => {
        if (!isScanOpen) {
            return
        }
        const handle = window.requestAnimationFrame(() => {
            scanInputRef.current?.focus()
        })
        return () => window.cancelAnimationFrame(handle)
    }, [isScanOpen])

    useEffect(() => {
        if (!selectedBook || searchMode !== "barcode") {
            return
        }

        const handleKeydown = (event: KeyboardEvent) => {
            const target = event.target as HTMLElement | null
            if (
                target &&
                (target.tagName === "INPUT" ||
                    target.tagName === "TEXTAREA" ||
                    target.isContentEditable)
            ) {
                return
            }

            if (/^\d$/.test(event.key)) {
                event.preventDefault()
                if (!isScanOpen) {
                    setIsScanOpen(true)
                    setScanValue(event.key)
                } else {
                    setScanValue((prev) => `${prev}${event.key}`)
                }
            } else if (event.key === "Enter" && isScanOpen) {
                event.preventDefault()
                void submitScan(scanValue)
            }
        }

        window.addEventListener("keydown", handleKeydown)
        return () => {
            window.removeEventListener("keydown", handleKeydown)
        }
    }, [isScanOpen, scanValue, searchMode, selectedBook, submitScan])

    const extraEntries =
        selectedBook?.extra && Object.keys(selectedBook.extra).length > 0
            ? Object.entries(selectedBook.extra)
            : []

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
                    <SearchBar
                        onSearch={handleSearch}
                        maxLength={MAX_QUERY_LENGTH}
                        onLimitExceeded={(limit) =>
                            setError(
                                `Превышена допустимая длина запроса (${limit} символов).`
                            )
                        }
                        isLoading={isLoading}
                    />
                    <button
                        className="primary-button"
                        type="button"
                        onClick={handleOpenAdd}
                    >
                        Добавить книгу
                    </button>
                    <button
                        className="ghost-button"
                        type="button"
                        onClick={handleOpenScan}
                    >
                        Поиск по штрихкоду
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
                            Поиск: "{truncateLabel(detailsQuery, MAX_QUERY_LABEL)}"
                        </span>
                    )}
                    {!detailsQuery && scanHighlight && (
                        <span className="query-chip">
                            Штрихкод: "{truncateLabel(scanHighlight, MAX_QUERY_LABEL)}"
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
                            <div className="modal-inline-actions">
                                {searchMode === "barcode" && (
                                    <button
                                        className="ghost-button"
                                        type="button"
                                        disabled
                                        title="Редактирование недоступно без API обновления"
                                    >
                                        Изменить
                                    </button>
                                )}
                                <button
                                    className="ghost-button"
                                    type="button"
                                    onClick={() => setSelectedBook(null)}
                                >
                                    Закрыть
                                </button>
                            </div>
                        </header>
                        <div className="modal-body">
                            <div className="detail-row">
                                <span>Автор</span>
                                <strong>{renderValue(selectedBook.author)}</strong>
                            </div>
                            <div className="detail-row">
                                <span>Штрихкод</span>
                                <strong>{renderBarcodeValue(selectedBook.barcode)}</strong>
                            </div>
                            <div className="detail-row">
                                <span>Заводской штрихкод</span>
                                <strong>
                                    {renderBarcodeValue(
                                        selectedBook.factory_barcode
                                    )}
                                </strong>
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
                            {extraEntries.length > 0 && (
                                <div className="extra-block">
                                    <p className="extra-title">Дополнительные поля</p>
                                    {extraEntries.map(([key, value]) => (
                                        <div className="detail-row" key={key}>
                                            <span>{key}</span>
                                            <strong>{renderExtraValue(value)}</strong>
                                        </div>
                                    ))}
                                </div>
                            )}
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
                <form
                    className="modal-body"
                    onSubmit={(event) => event.preventDefault()}
                    onKeyDown={(event) => {
                        if (
                            event.key === "Enter" &&
                            (event.target as HTMLElement).tagName !== "TEXTAREA"
                        ) {
                            event.preventDefault()
                        }
                    }}
                >
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
                                Заводской штрихкод
                                <input
                                    inputMode="numeric"
                                    value={draft.factoryBarcode}
                                    onChange={(event) =>
                                        setDraft((prev) => ({
                                            ...prev,
                                            factoryBarcode: event.target.value,
                                        }))
                                    }
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
                            <div className="extra-fields">
                                <div className="extra-header">
                                    <span>Дополнительные поля</span>
                                    <button
                                        className="ghost-button"
                                        type="button"
                                        onClick={() =>
                                            setDraft((prev) => ({
                                                ...prev,
                                                extra: [
                                                    ...prev.extra,
                                                    {
                                                        id: createExtraId(),
                                                        key: "",
                                                        value: "",
                                                    },
                                                ],
                                            }))
                                        }
                                    >
                                        Добавить поле
                                    </button>
                                </div>
                                {draft.extra.map((field, index) => (
                                    <div className="extra-row" key={field.id}>
                                        <input
                                            className="extra-input"
                                            placeholder="Название"
                                            value={field.key}
                                            onChange={(event) =>
                                                setDraft((prev) => {
                                                    const next = [...prev.extra]
                                                    next[index] = {
                                                        ...next[index],
                                                        key: event.target.value,
                                                    }
                                                    return {...prev, extra: next}
                                                })
                                            }
                                        />
                                        <textarea
                                            className="extra-input extra-textarea"
                                            placeholder="Значение"
                                            value={field.value}
                                            onChange={(event) =>
                                                setDraft((prev) => {
                                                    const next = [...prev.extra]
                                                    next[index] = {
                                                        ...next[index],
                                                        value: event.target.value,
                                                    }
                                                    return {...prev, extra: next}
                                                })
                                            }
                                        />
                                        <button
                                            className="ghost-button"
                                            type="button"
                                            onClick={() =>
                                                setDraft((prev) => ({
                                                    ...prev,
                                                    extra: prev.extra.filter(
                                                        (_, itemIndex) =>
                                                            itemIndex !== index
                                                    ),
                                                }))
                                            }
                                            disabled={draft.extra.length === 1}
                                        >
                                            Удалить
                                        </button>
                                    </div>
                                ))}
                            </div>
                            {addError && <p className="error-banner">{addError}</p>}
                            <div className="modal-actions">
                                <button
                                    className="primary-button"
                                    type="button"
                                    disabled={isAddSaving}
                                    onClick={() => {
                                        void handleAddSubmit()
                                    }}
                                >
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

            {isScanOpen && (
                <div
                    className="modal-backdrop"
                    role="dialog"
                    aria-modal="true"
                    onClick={(event) => {
                        if (event.target === event.currentTarget) {
                            setIsScanOpen(false)
                        }
                    }}
                >
                    <div className="modal-card">
                        <header className="modal-header">
                            <div>
                                <p className="modal-eyebrow">Сканирование</p>
                                <h3>Поиск по штрихкоду</h3>
                            </div>
                            <button
                                className="ghost-button"
                                type="button"
                                onClick={() => setIsScanOpen(false)}
                            >
                                Закрыть
                            </button>
                        </header>
                        <form className="modal-body" onSubmit={handleScanSubmit}>
                            <label className="field">
                                Штрихкод
                                <input
                                    ref={scanInputRef}
                                    inputMode="numeric"
                                    value={scanValue}
                                    onChange={(event) =>
                                        setScanValue(event.target.value)
                                    }
                                />
                            </label>
                            {scanMessage && (
                                <p className="status-line">{scanMessage}</p>
                            )}
                            <div className="modal-actions">
                                <button
                                    className="primary-button"
                                    type="submit"
                                    disabled={isScanLoading}
                                >
                                    Найти
                                </button>
                                <button
                                    className="ghost-button"
                                    type="button"
                                    onClick={() => setIsScanOpen(false)}
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

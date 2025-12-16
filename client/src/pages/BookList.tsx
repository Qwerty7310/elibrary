import {useState} from "react"
import type {Book} from "../types/book"
import {scanBook, searchBooks} from "../api/books"
import {SearchBar} from "../components/SearchBar"

export function BookList() {
    const [books, setBooks] = useState<Book[]>([])
    const [error, setError] = useState<string | null>(null)

    async function handleSearch(value: string) {
        setError(null)

        try {
            // сначала пробуем как скан
            const book = await scanBook(value)
            setBooks([book])
        } catch {
            try {
                // если не нашли — обычный поиск
                const result = await searchBooks(value)
                setBooks(result)
            } catch {
                setError("Ничего не найдено")
            }
        }
    }

    return (
        <div>
            <h1>Библиотека</h1>
            <SearchBar onSearch={handleSearch}/>

            {error && <p>{error}</p>}

            <ul>
                {books.map((b) => (
                    <li key={b.id}>
                        {b.title} — {b.author}
                    </li>
                ))}
            </ul>
        </div>
    )
}

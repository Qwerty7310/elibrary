import type {Book} from "../types/book"

const API_URL = "http://localhost:8080"

export async function searchBooks(query: string): Promise<Book[]> {
    const res = await fetch(
        `${API_URL}/books/search?q=${encodeURIComponent(query)}`
    )
    if (!res.ok) {
        throw new Error("search failed")
    }
    return res.json()
}

export async function scanBook(value: string): Promise<Book> {
    const res = await fetch(`${API_URL}/scan/${encodeURIComponent(value)}`)
    if (!res.ok) {
        throw new Error("not found")
    }
    return res.json()
}
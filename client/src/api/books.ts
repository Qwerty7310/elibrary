import type {Book} from "../types/book"

// const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080"

// Для продакшена используем путь через Nginx
const API_URL = window.location.hostname === 'localhost'
    ? 'http://localhost:8080'  // для локальной разработки
    : '/elibrary/api'           // для продакшена через Nginx

export async function searchBooks(query: string): Promise<Book[]> {
    const res = await fetch(
        `${API_URL}/books/search?q=${encodeURIComponent(query)}`
    )
    if (!res.ok) {
        throw new Error(`search failed: ${res.status}`)
    }
    const data = await res.json()
    return data.books || []
}

export async function scanBook(value: string): Promise<Book> {
    const res = await fetch(`${API_URL}/scan/${encodeURIComponent(value)}`)
    if (!res.ok) {
        throw new Error("not found")
    }
    return res.json()
}
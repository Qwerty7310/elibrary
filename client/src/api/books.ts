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

export async function scanBooks(value: string): Promise<Book[]> {
    const res = await fetch(`${API_URL}/scan/${encodeURIComponent(value)}`)
    if (!res.ok) {
        const error = new Error("scan failed") as Error & {status?: number}
        error.status = res.status
        throw error
    }
    return res.json()
}

export async function createBook(payload: {
    title: string
    author: string
    publisher?: string
    year?: number
    location?: string
    factory_barcode?: string
    extra?: Record<string, unknown>
}): Promise<Book> {
    const res = await fetch(`${API_URL}/books`, {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify(payload),
    })

    if (!res.ok) {
        const error = new Error("create failed") as Error & {status?: number}
        error.status = res.status
        throw error
    }

    const data = await res.json()
    return data.book
}

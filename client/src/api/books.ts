import type {
    BookInternal,
    BookPublic,
    BookWorkInput,
} from "../types/library"
import {requestJson} from "./http"

type BookListResponse<T> = {
    items: T[]
    count: number
}

const BOOKS_PAGE_LIMIT = 200

async function fetchAllPages<T>(
    fetchPage: (limit: number, offset: number) => Promise<BookListResponse<T>>
): Promise<T[]> {
    const all: T[] = []
    let offset = 0
    while (true) {
        const data = await fetchPage(BOOKS_PAGE_LIMIT, offset)
        const items = data.items ?? []
        all.push(...items)
        if (items.length < BOOKS_PAGE_LIMIT) {
            break
        }
        offset += items.length
    }
    return all
}

export async function searchBooksPublic(query: string): Promise<BookPublic[]> {
    const trimmed = query.trim()
    return fetchAllPages((limit, offset) => {
        const params = new URLSearchParams()
        if (trimmed) {
            params.set("q", trimmed)
        }
        params.set("limit", String(limit))
        params.set("offset", String(offset))
        return requestJson<BookListResponse<BookPublic>>(
            `/books/public?${params.toString()}`
        )
    })
}

export async function searchBooksInternal(query: string): Promise<BookInternal[]> {
    const trimmed = query.trim()
    return fetchAllPages((limit, offset) => {
        const params = new URLSearchParams()
        if (trimmed) {
            params.set("q", trimmed)
        }
        params.set("limit", String(limit))
        params.set("offset", String(offset))
        return requestJson<BookListResponse<BookInternal>>(
            `/books/internal?${params.toString()}`
        )
    })
}

export async function createBook(payload: {
    book: {
        title: string
        publisher_id?: string
        year?: number
        description?: string
        location_id?: string
        factory_barcode?: string
        extra?: Record<string, unknown>
    }
    works: BookWorkInput[]
}): Promise<BookPublic> {
    return requestJson<BookPublic>("/admin/books", {
        method: "POST",
        body: JSON.stringify(payload),
    })
}

export async function updateBook(
    id: string,
    payload: {
        title?: string
        publisher_id?: string
        year?: number
        description?: string
        location_id?: string
        factory_barcode?: string
        extra?: Record<string, unknown>
        works?: BookWorkInput[]
    }
): Promise<void> {
    return requestJson<void>(`/admin/books/${encodeURIComponent(id)}`, {
        method: "PUT",
        body: JSON.stringify(payload),
    })
}

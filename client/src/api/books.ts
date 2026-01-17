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

export async function searchBooksPublic(query: string): Promise<BookPublic[]> {
    const params = new URLSearchParams()
    if (query.trim()) {
        params.set("q", query.trim())
    }
    const data = await requestJson<BookListResponse<BookPublic>>(
        `/books/public?${params.toString()}`
    )
    return data.items ?? []
}

export async function searchBooksInternal(query: string): Promise<BookInternal[]> {
    const params = new URLSearchParams()
    if (query.trim()) {
        params.set("q", query.trim())
    }
    const data = await requestJson<BookListResponse<BookInternal>>(
        `/books/internal?${params.toString()}`
    )
    return data.items ?? []
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

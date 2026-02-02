import type {Author} from "../types/library"
import {requestJson} from "./http"

export function createAuthor(payload: {
    last_name: string
    first_name?: string
    middle_name?: string
    birth_date?: string
    death_date?: string
    bio?: string
    photo_url?: string
}) {
    return requestJson<Author>("/admin/authors", {
        method: "POST",
        body: JSON.stringify({
            ...payload,
            birth_date: normalizeDate(payload.birth_date),
            death_date: normalizeDate(payload.death_date),
        }),
    })
}

export function updateAuthor(id: string, payload: {
    last_name?: string
    first_name?: string
    middle_name?: string
    birth_date?: string
    death_date?: string
    bio?: string
    photo_url?: string
}) {
    return requestJson<void>(`/admin/authors/${encodeURIComponent(id)}`, {
        method: "PUT",
        body: JSON.stringify({
            ...payload,
            birth_date: normalizeDate(payload.birth_date),
            death_date: normalizeDate(payload.death_date),
        }),
    })
}

export function getAuthorByID(id: string) {
    return requestJson<Author>(`/authors/${encodeURIComponent(id)}`)
}

function normalizeDate(value?: string) {
    if (!value) {
        return undefined
    }
    return `${value}T00:00:00Z`
}

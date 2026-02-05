import type {WorkDetailed} from "../types/library"
import {requestJson} from "./http"

export function createWork(payload: {
    work: {
        title: string
        description?: string
        year?: number
    }
    authors: string[]
}) {
    return requestJson<WorkDetailed>("/admin/works", {
        method: "POST",
        body: JSON.stringify(payload),
    })
}

export function updateWork(
    id: string,
    payload: {
        title?: string
        description?: string
        year?: number
        authors?: string[]
    }
) {
    return requestJson<void>(`/admin/works/${encodeURIComponent(id)}`, {
        method: "PUT",
        body: JSON.stringify(payload),
    })
}

export function getWorkByID(id: string) {
    return requestJson<WorkDetailed>(`/works/${encodeURIComponent(id)}`)
}

export function deleteWork(id: string) {
    return requestJson<void>(`/admin/works/${encodeURIComponent(id)}`, {
        method: "DELETE",
    })
}

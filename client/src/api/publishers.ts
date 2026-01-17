import type {Publisher} from "../types/library"
import {requestJson} from "./http"

export function createPublisher(payload: {
    name: string
    logo_url?: string
    web_url?: string
}) {
    return requestJson<Publisher>("/admin/publishers", {
        method: "POST",
        body: JSON.stringify(payload),
    })
}

export function getPublisherByID(id: string) {
    return requestJson<Publisher>(`/publishers/${encodeURIComponent(id)}`)
}


import type {AuthorSummary, Publisher, WorkShort} from "../types/library"
import {requestJson} from "./http"

export async function getAuthorsReference() {
    const data = await requestJson<AuthorSummary[] | null>("/reference/authors")
    return data ?? []
}

export async function getWorksReference() {
    const data = await requestJson<WorkShort[] | null>("/reference/works")
    return data ?? []
}

export async function getPublishersReference() {
    const data = await requestJson<Publisher[] | null>("/reference/publishers")
    return data ?? []
}

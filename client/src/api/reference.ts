import type {AuthorSummary, Publisher, WorkShort} from "../types/library"
import {requestJson} from "./http"

export function getAuthorsReference() {
    return requestJson<AuthorSummary[]>("/reference/authors")
}

export function getWorksReference() {
    return requestJson<WorkShort[]>("/reference/works")
}

export function getPublishersReference() {
    return requestJson<Publisher[]>("/reference/publishers")
}


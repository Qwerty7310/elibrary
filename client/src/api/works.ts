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


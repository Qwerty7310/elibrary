import type {LocationEntity} from "../types/library"
import {ApiError, requestJson} from "./http"

export function createLocation(payload: {
    parent_id?: string
    type: string
    name: string
    address?: string
    description?: string
}) {
    return requestJson<LocationEntity>("/admin/locations", {
        method: "POST",
        body: JSON.stringify(payload),
    })
}

export function updateLocation(id: string, payload: {name?: string}) {
    return requestJson<void>(`/admin/locations/${encodeURIComponent(id)}`, {
        method: "PUT",
        body: JSON.stringify(payload),
    })
}

export function deleteLocation(id: string) {
    return requestJson<void>(`/admin/locations/${encodeURIComponent(id)}`, {
        method: "DELETE",
    })
}

export function getLocationsByType(type: string) {
    return requestJson<LocationEntity[] | null>(
        `/locations/type/${encodeURIComponent(type)}`
    )
}

export async function getLocationChildren(parentId: string, type: string) {
    try {
        return await requestJson<LocationEntity[] | null>(
            `/locations/child/${encodeURIComponent(
                parentId
            )}/${encodeURIComponent(type)}`
        )
    } catch (error) {
        const apiError = error as ApiError
        if (apiError.status === 404) {
            return []
        }
        throw error
    }
}

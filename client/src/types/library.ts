export type Role = {
    id: number
    code: string
    name: string
}

export type User = {
    id: string
    login: string
    first_name: string
    last_name?: string
    middle_name?: string
    email?: string
    is_active: boolean
    roles: Role[]
}

export type AuthorSummary = {
    id: string
    last_name: string
    first_name?: string
    middle_name?: string
    photo_url?: string
}

export type Author = AuthorSummary & {
    birth_date?: string
    death_date?: string
    bio?: string
    photo_url?: string
}

export type Publisher = {
    id: string
    name: string
    logo_url?: string
    web_url?: string
}

export type WorkShort = {
    id: string
    title: string
    authors: AuthorSummary[]
    year?: number
}

export type WorkDetailed = {
    id: string
    title: string
    description?: string
    year?: number
    authors: AuthorSummary[]
}

export type BookWorkInput = {
    work_id: string
    position?: number | null
}

export type BookBase = {
    id: string
    title: string
    barcode: string
    factory_barcode?: string
    publisher?: Publisher
    works?: WorkShort[]
    year?: number
    description?: string
    extra?: Record<string, unknown>
    created_at: string
    updated_at: string
}

export type BookLocation = {
    shelf_id: string
    shelf_name: string
    cabinet_id: string
    cabinet_name: string
    room_id: string
    room_name: string
    building_id: string
    building_name: string
    address: string
}

export type BookPublic = BookBase

export type BookInternal = BookBase & {
    location?: BookLocation
}

export type LocationEntity = {
    id: string
    parent_id?: string
    type: string
    name: string
    barcode: string
    address?: string
    description?: string
    created_at: string
    updated_at: string
}

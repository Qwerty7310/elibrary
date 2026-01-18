import {useEffect, useMemo, useState} from "react"
import "./App.css"
import {loginUser} from "./api/auth"
import {createBook, searchBooksInternal, searchBooksPublic} from "./api/books"
import {createAuthor, getAuthorByID} from "./api/authors"
import {createPublisher} from "./api/publishers"
import {createWork} from "./api/works"
import {
    createLocation,
    getLocationChildren,
    getLocationsByType,
} from "./api/locations"
import {
    getAuthorsReference,
    getPublishersReference,
    getWorksReference,
} from "./api/reference"
import {ApiError, setToken} from "./api/http"
import {getUserByID, updateUser} from "./api/users"
import {SearchBar} from "./components/SearchBar"
import {SafeImage} from "./components/SafeImage"
import type {
    Author,
    AuthorSummary,
    BookInternal,
    BookPublic,
    BookWorkInput,
    LocationEntity,
    Publisher,
    User,
    WorkShort,
} from "./types/library"

type TabKey =
    | "books"
    | "works"
    | "authors"
    | "publishers"
    | "locations"
    | "profile"

type BookDraft = {
    title: string
    publisherId: string
    year: string
    description: string
    locationId: string
    factoryBarcode: string
    workIds: string[]
}

type WorkDraft = {
    title: string
    description: string
    year: string
    authorIds: string[]
}

type AuthorDraft = {
    lastName: string
    firstName: string
    middleName: string
    birthDate: string
    deathDate: string
    bio: string
    photoUrl: string
}

type PublisherDraft = {
    name: string
    logoUrl: string
    webUrl: string
}

type LocationDraft = {
    parentId: string
    type: string
    name: string
    address: string
    description: string
    lockParent: boolean
    lockType: boolean
}

const emptyBookDraft: BookDraft = {
    title: "",
    publisherId: "",
    year: "",
    description: "",
    locationId: "",
    factoryBarcode: "",
    workIds: [],
}

const emptyWorkDraft: WorkDraft = {
    title: "",
    description: "",
    year: "",
    authorIds: [],
}

const emptyAuthorDraft: AuthorDraft = {
    lastName: "",
    firstName: "",
    middleName: "",
    birthDate: "",
    deathDate: "",
    bio: "",
    photoUrl: "",
}

const emptyPublisherDraft: PublisherDraft = {
    name: "",
    logoUrl: "",
    webUrl: "",
}

const emptyLocationDraft: LocationDraft = {
    parentId: "",
    type: "",
    name: "",
    address: "",
    description: "",
    lockParent: false,
    lockType: false,
}

function parseJwt(token: string) {
    try {
        const base64 = token.split(".")[1]
        const normalized = base64.replace(/-/g, "+").replace(/_/g, "/")
        const padded = normalized.padEnd(
            normalized.length + ((4 - (normalized.length % 4)) % 4),
            "="
        )
        const json = atob(padded)
        return JSON.parse(json) as {sub?: string}
    } catch {
        return null
    }
}

function getAuthorName(author: AuthorSummary) {
    return [author.last_name, author.first_name, author.middle_name]
        .filter(Boolean)
        .join(" ")
}

function formatLocation(location?: BookInternal["location"]) {
    if (!location) {
        return "—"
    }
    const parts = [
        location.building_name,
        location.room_name,
        location.cabinet_name,
        location.shelf_name,
    ].filter(Boolean)
    const address = location.address ? `, ${location.address}` : ""
    return `${parts.join(" · ")}${address}`
}

function getCoverUrl(book: BookPublic) {
    const extra = book.extra ?? {}
    const cover = extra.cover_url
    return typeof cover === "string" ? cover : null
}

export default function App() {
    const [activeTab, setActiveTab] = useState<TabKey>("books")
    const [token, setAuthToken] = useState<string | null>(
        localStorage.getItem("auth_token")
    )
    const [loginName, setLoginName] = useState(
        localStorage.getItem("login_name") ?? ""
    )
    const [user, setUser] = useState<User | null>(null)
    const [isAdmin, setIsAdmin] = useState(false)
    const [authError, setAuthError] = useState<string | null>(null)
    const [isLoginOpen, setIsLoginOpen] = useState(false)
    const [loginDraft, setLoginDraft] = useState({
        login: "",
        password: "",
    })

    const [authors, setAuthors] = useState<AuthorSummary[]>([])
    const [works, setWorks] = useState<WorkShort[]>([])
    const [publishers, setPublishers] = useState<Publisher[]>([])
    const [locationByType, setLocationByType] = useState<
        Record<string, LocationEntity[]>
    >({})
    const [locationChildren, setLocationChildren] = useState<
        Record<string, LocationEntity[]>
    >({})
    const [expandedLocations, setExpandedLocations] = useState<Set<string>>(
        () => new Set()
    )
    const [locationLoading, setLocationLoading] = useState<Set<string>>(
        () => new Set()
    )
    const [locationsError, setLocationsError] = useState<string | null>(null)
    const [locationsLoaded, setLocationsLoaded] = useState(false)

    const [booksQuery, setBooksQuery] = useState("")
    const [books, setBooks] = useState<BookPublic[]>([])
    const [booksError, setBooksError] = useState<string | null>(null)
    const [booksLoading, setBooksLoading] = useState(false)

    const [workQuery, setWorkQuery] = useState("")
    const [selectedWork, setSelectedWork] = useState<WorkShort | null>(null)
    const [selectedWorkDetail, setSelectedWorkDetail] =
        useState<WorkShort | null>(null)
    const [isWorkInfoOpen, setIsWorkInfoOpen] = useState(false)
    const [workBooks, setWorkBooks] = useState<BookPublic[]>([])
    const [workBooksLoading, setWorkBooksLoading] = useState(false)

    const [authorQuery, setAuthorQuery] = useState("")
    const [selectedAuthor, setSelectedAuthor] = useState<Author | null>(null)
    const [isAuthorInfoOpen, setIsAuthorInfoOpen] = useState(false)
    const [authorBooks, setAuthorBooks] = useState<BookPublic[]>([])
    const [authorBooksLoading, setAuthorBooksLoading] = useState(false)

    const [publisherQuery, setPublisherQuery] = useState("")

    const [isBookModalOpen, setIsBookModalOpen] = useState(false)
    const [bookDraft, setBookDraft] = useState<BookDraft>(emptyBookDraft)
    const [bookError, setBookError] = useState<string | null>(null)
    const [bookSaving, setBookSaving] = useState(false)
    const [coverPreview, setCoverPreview] = useState<string | null>(null)
    const [isWorksPickerOpen, setIsWorksPickerOpen] = useState(true)
    const [selectedBuildingId, setSelectedBuildingId] = useState("")
    const [selectedRoomId, setSelectedRoomId] = useState("")
    const [selectedCabinetId, setSelectedCabinetId] = useState("")
    const [selectedShelfId, setSelectedShelfId] = useState("")

    const [isWorkModalOpen, setIsWorkModalOpen] = useState(false)
    const [workDraft, setWorkDraft] = useState<WorkDraft>(emptyWorkDraft)
    const [workError, setWorkError] = useState<string | null>(null)
    const [workSaving, setWorkSaving] = useState(false)

    const [isAuthorModalOpen, setIsAuthorModalOpen] = useState(false)
    const [authorDraft, setAuthorDraft] = useState<AuthorDraft>(emptyAuthorDraft)
    const [authorError, setAuthorError] = useState<string | null>(null)
    const [authorSaving, setAuthorSaving] = useState(false)

    const [isPublisherModalOpen, setIsPublisherModalOpen] = useState(false)
    const [publisherDraft, setPublisherDraft] =
        useState<PublisherDraft>(emptyPublisherDraft)
    const [publisherError, setPublisherError] = useState<string | null>(null)
    const [publisherSaving, setPublisherSaving] = useState(false)

    const [isLocationModalOpen, setIsLocationModalOpen] = useState(false)
    const [locationDraft, setLocationDraft] =
        useState<LocationDraft>(emptyLocationDraft)
    const [locationError, setLocationError] = useState<string | null>(null)
    const [locationSaving, setLocationSaving] = useState(false)

    const [profileDraft, setProfileDraft] = useState({
        login: "",
        first_name: "",
        last_name: "",
        middle_name: "",
        email: "",
        password: "",
        is_active: true,
    })
    const [profileError, setProfileError] = useState<string | null>(null)
    const [profileSaving, setProfileSaving] = useState(false)

    useEffect(() => {
        if (!token) {
            setIsLoginOpen(true)
            setUser(null)
            setIsAdmin(false)
            return
        }

        const claims = parseJwt(token)
        const userId = claims?.sub
        if (!userId) {
            setAuthToken(null)
            setToken(null)
            return
        }

        ;(async () => {
            try {
                const current = await getUserByID(userId)
                setUser(current)
                setIsAdmin(current.roles.some((role) => role.code === "admin"))
                setProfileDraft({
                    login: current.login,
                    first_name: current.first_name,
                    last_name: current.last_name ?? "",
                    middle_name: current.middle_name ?? "",
                    email: current.email ?? "",
                    password: "",
                    is_active: current.is_active,
                })
            } catch (err) {
                const apiErr = err as ApiError
                if (apiErr.status === 403 || apiErr.status === 404) {
                    const fallbackUser = {
                        id: userId,
                        login: loginName || "user",
                        first_name: "",
                        last_name: "",
                        middle_name: "",
                        email: "",
                        is_active: true,
                        roles: [],
                    }
                    setUser(fallbackUser)
                    setIsAdmin(false)
                    setProfileDraft({
                        login: fallbackUser.login,
                        first_name: "",
                        last_name: "",
                        middle_name: "",
                        email: "",
                        password: "",
                        is_active: true,
                    })
                    return
                }
                setAuthError("Не удалось загрузить профиль")
            }
        })()
    }, [token, loginName])

    useEffect(() => {
        if (!token) {
            return
        }
        ;(async () => {
            try {
                const [authorsData, worksData, publishersData] =
                    await Promise.all([
                        getAuthorsReference(),
                        getWorksReference(),
                        getPublishersReference(),
                    ])
                setAuthors(authorsData)
                setWorks(worksData ?? [])
                setPublishers(publishersData)
            } catch {
                setAuthError("Не удалось загрузить справочники")
            }
        })()
    }, [token])

    useEffect(() => {
        if (isBookModalOpen && token) {
            loadBuildingLocations()
            setSelectedBuildingId("")
            setSelectedRoomId("")
            setSelectedCabinetId("")
            setSelectedShelfId("")
            setBookDraft((prev) => ({...prev, locationId: ""}))
            setIsWorksPickerOpen(true)
        }
    }, [isBookModalOpen, token])

    const filteredWorks = useMemo(() => {
        const needle = workQuery.trim().toLowerCase()
        if (!needle) {
            return works
        }
        return works.filter((work) =>
            work.title.toLowerCase().includes(needle)
        )
    }, [workQuery, works])

    const filteredAuthors = useMemo(() => {
        const needle = authorQuery.trim().toLowerCase()
        if (!needle) {
            return authors
        }
        return authors.filter((author) =>
            getAuthorName(author).toLowerCase().includes(needle)
        )
    }, [authorQuery, authors])

    const filteredPublishers = useMemo(() => {
        const needle = publisherQuery.trim().toLowerCase()
        if (!needle) {
            return publishers
        }
        return publishers.filter((publisher) =>
            publisher.name.toLowerCase().includes(needle)
        )
    }, [publisherQuery, publishers])

    const buildingLocations = useMemo(
        () => locationByType.building ?? [],
        [locationByType]
    )

    const roomLocations = useMemo(
        () => (selectedBuildingId ? locationChildren[selectedBuildingId] ?? [] : []),
        [locationChildren, selectedBuildingId]
    )

    const cabinetLocations = useMemo(
        () => (selectedRoomId ? locationChildren[selectedRoomId] ?? [] : []),
        [locationChildren, selectedRoomId]
    )

    const shelfLocationsForSelection = useMemo(
        () => (selectedCabinetId ? locationChildren[selectedCabinetId] ?? [] : []),
        [locationChildren, selectedCabinetId]
    )

    function getParentType(type: string) {
        if (type === "room") return "building"
        if (type === "cabinet") return "room"
        if (type === "shelf") return "cabinet"
        return null
    }

    async function ensureLocationsByType(type: string) {
        if (locationByType[type]) {
            return
        }
        try {
            const data = await getLocationsByType(type)
            setLocationByType((prev) => ({
                ...prev,
                [type]: data ?? [],
            }))
            setLocationsError(null)
        } catch {
            setLocationsError("Не удалось загрузить локации")
        }
    }

    async function loadBuildingLocations() {
        if (locationsLoaded) {
            return
        }
        try {
            const data = await getLocationsByType("building")
            setLocationByType((prev) => ({
                ...prev,
                building: data ?? [],
            }))
            setLocationsLoaded(true)
            setLocationsError(null)
        } catch {
            setLocationsError("Не удалось загрузить локации")
        }
    }

    async function handleLoginSubmit(event: React.FormEvent) {
        event.preventDefault()
        setAuthError(null)
        try {
            const tokenValue = await loginUser(
                loginDraft.login.trim(),
                loginDraft.password
            )
            setAuthToken(tokenValue)
            setToken(tokenValue)
            localStorage.setItem("login_name", loginDraft.login.trim())
            setLoginName(loginDraft.login.trim())
            setIsLoginOpen(false)
            setLoginDraft({login: "", password: ""})
        } catch {
            setAuthError("Не удалось войти. Проверьте логин и пароль.")
        }
    }

    function handleLogout() {
        setAuthToken(null)
        setToken(null)
        setUser(null)
        setIsAdmin(false)
        localStorage.removeItem("login_name")
        setLoginName("")
        setActiveTab("books")
        setBooks([])
    }

    async function handleBookSearch(value: string) {
        if (!token) {
            setIsLoginOpen(true)
            return
        }
        setBooksQuery(value)
        setBooksError(null)
        if (!value.trim()) {
            setBooks([])
            return
        }
        setBooksLoading(true)
        try {
            const data = isAdmin
                ? await searchBooksInternal(value)
                : await searchBooksPublic(value)
            setBooks(data)
            if (data.length === 0) {
                setBooksError("Ничего не найдено")
            }
        } catch {
            setBooksError("Не удалось выполнить поиск")
        } finally {
            setBooksLoading(false)
        }
    }

    async function handleWorkSelect(work: WorkShort) {
        setSelectedWork(work)
        setSelectedWorkDetail(work)
        setIsWorkInfoOpen(true)
        setWorkBooks([])
        setWorkBooksLoading(true)
        try {
            const data = isAdmin
                ? await searchBooksInternal(work.title)
                : await searchBooksPublic(work.title)
            setWorkBooks(data)
        } catch {
            setWorkBooks([])
        } finally {
            setWorkBooksLoading(false)
        }
    }

    async function handleAuthorSelect(author: AuthorSummary) {
        setSelectedAuthor(null)
        setIsAuthorInfoOpen(true)
        setAuthorBooks([])
        setAuthorBooksLoading(true)
        try {
            const [details, data] = await Promise.all([
                getAuthorByID(author.id),
                isAdmin
                    ? searchBooksInternal(getAuthorName(author))
                    : searchBooksPublic(getAuthorName(author)),
            ])
            setSelectedAuthor(details)
            setAuthorBooks(data)
        } catch {
            setAuthorBooks([])
        } finally {
            setAuthorBooksLoading(false)
        }
    }

    async function handleCreateBook() {
        if (!bookDraft.title.trim()) {
            setBookError("Название книги обязательно")
            return
        }
        if (bookDraft.workIds.length === 0) {
            setBookError("Выберите хотя бы одно произведение")
            return
        }
        setBookSaving(true)
        setBookError(null)
        const worksPayload: BookWorkInput[] = bookDraft.workIds.map(
            (workId, index) => ({
                work_id: workId,
                position: index + 1,
            })
        )
        const payload = {
            book: {
                title: bookDraft.title.trim(),
                publisher_id: bookDraft.publisherId || undefined,
                year: bookDraft.year ? Number(bookDraft.year) : undefined,
                description: bookDraft.description.trim() || undefined,
                location_id: bookDraft.locationId || undefined,
                factory_barcode: bookDraft.factoryBarcode.trim() || undefined,
                extra: undefined,
            },
            works: worksPayload,
        }
        try {
            await createBook(payload)
            setBookDraft(emptyBookDraft)
            setCoverPreview(null)
            setIsBookModalOpen(false)
            if (booksQuery.trim()) {
                await handleBookSearch(booksQuery)
            }
        } catch {
            setBookError("Не удалось создать книгу")
        } finally {
            setBookSaving(false)
        }
    }

    async function handleCreateWork() {
        if (!workDraft.title.trim()) {
            setWorkError("Название произведения обязательно")
            return
        }
        setWorkSaving(true)
        setWorkError(null)
        try {
            const created = await createWork({
                work: {
                    title: workDraft.title.trim(),
                    description: workDraft.description.trim() || undefined,
                    year: workDraft.year ? Number(workDraft.year) : undefined,
                },
                authors: workDraft.authorIds,
            })
            const selectedAuthors = authors.filter((author) =>
                workDraft.authorIds.includes(author.id)
            )
            const year = workDraft.year ? Number(workDraft.year) : undefined
            setWorks((prev) => [
                {
                    id: created.id,
                    title: created.title,
                    authors: selectedAuthors,
                    year,
                },
                ...prev,
            ])
            setWorkDraft(emptyWorkDraft)
            setIsWorkModalOpen(false)
            if (isBookModalOpen) {
                setBookDraft((prev) => ({
                    ...prev,
                    workIds: prev.workIds.includes(created.id)
                        ? prev.workIds
                        : [...prev.workIds, created.id],
                }))
            }
        } catch {
            setWorkError("Не удалось создать произведение")
        } finally {
            setWorkSaving(false)
        }
    }

    async function handleCreateAuthor() {
        if (!authorDraft.lastName.trim()) {
            setAuthorError("Фамилия обязательна")
            return
        }
        setAuthorSaving(true)
        setAuthorError(null)
        try {
            const created = await createAuthor({
                last_name: authorDraft.lastName.trim(),
                first_name: authorDraft.firstName.trim() || undefined,
                middle_name: authorDraft.middleName.trim() || undefined,
                birth_date: authorDraft.birthDate || undefined,
                death_date: authorDraft.deathDate || undefined,
                bio: authorDraft.bio.trim() || undefined,
                photo_url: authorDraft.photoUrl.trim() || undefined,
            })
            const summary = {
                id: created.id,
                last_name: created.last_name,
                first_name: created.first_name,
                middle_name: created.middle_name,
            }
            setAuthors((prev) => [summary, ...prev])
            setAuthorDraft(emptyAuthorDraft)
            setIsAuthorModalOpen(false)
            if (isWorkModalOpen) {
                setWorkDraft((prev) => ({
                    ...prev,
                    authorIds: prev.authorIds.includes(created.id)
                        ? prev.authorIds
                        : [...prev.authorIds, created.id],
                }))
            }
        } catch {
            setAuthorError("Не удалось создать автора")
        } finally {
            setAuthorSaving(false)
        }
    }

    async function handleCreatePublisher() {
        if (!publisherDraft.name.trim()) {
            setPublisherError("Название издательства обязательно")
            return
        }
        setPublisherSaving(true)
        setPublisherError(null)
        try {
            const created = await createPublisher({
                name: publisherDraft.name.trim(),
                logo_url: publisherDraft.logoUrl.trim() || undefined,
                web_url: publisherDraft.webUrl.trim() || undefined,
            })
            setPublishers((prev) => [created, ...prev])
            setPublisherDraft(emptyPublisherDraft)
            setIsPublisherModalOpen(false)
            if (isBookModalOpen) {
                setBookDraft((prev) => ({
                    ...prev,
                    publisherId: created.id,
                }))
            }
        } catch {
            setPublisherError("Не удалось создать издательство")
        } finally {
            setPublisherSaving(false)
        }
    }

    async function handleCreateLocation() {
        if (!locationDraft.name.trim()) {
            setLocationError("Название локации обязательно")
            return
        }
        if (!locationDraft.type.trim()) {
            setLocationError("Тип локации обязателен")
            return
        }
        const parentType = getParentType(locationDraft.type.trim())
        if (parentType && !locationDraft.parentId) {
            setLocationError("Выберите родителя для локации")
            return
        }
        setLocationSaving(true)
        setLocationError(null)
        try {
            const created = await createLocation({
                parent_id: locationDraft.parentId || undefined,
                type: locationDraft.type.trim(),
                name: locationDraft.name.trim(),
                address: locationDraft.address.trim() || undefined,
                description: locationDraft.description.trim() || undefined,
            })
            setLocationByType((prev) => ({
                ...prev,
                [created.type]: [created, ...(prev[created.type] ?? [])],
            }))
            setLocationDraft(emptyLocationDraft)
            setIsLocationModalOpen(false)
            if (isBookModalOpen) {
                if (created.type === "building") {
                    setSelectedBuildingId(created.id)
                }
                if (created.type === "room") {
                    setSelectedRoomId(created.id)
                }
                if (created.type === "cabinet") {
                    setSelectedCabinetId(created.id)
                }
                if (created.type === "shelf") {
                    setSelectedShelfId(created.id)
                    setBookDraft((prev) => ({
                        ...prev,
                        locationId: created.id,
                    }))
                }
            }
            if (created.parent_id) {
                setLocationChildren((prev) => {
                    const current = prev[created.parent_id ?? ""] ?? []
                    return {
                        ...prev,
                        [created.parent_id ?? ""]: [created, ...current],
                    }
                })
            }
        } catch {
            setLocationError("Не удалось создать локацию")
        } finally {
            setLocationSaving(false)
        }
    }

    async function toggleLocation(location: LocationEntity) {
        const id = location.id
        const isExpanded = expandedLocations.has(id)
        if (isExpanded) {
            setExpandedLocations((prev) => {
                const next = new Set(prev)
                next.delete(id)
                return next
            })
            return
        }

        setExpandedLocations((prev) => new Set(prev).add(id))

        const childType = getChildType(location.type)
        if (!childType || locationChildren[id]) {
            return
        }

        setLocationLoading((prev) => new Set(prev).add(id))
        try {
            const data = await getLocationChildren(id, childType)
            setLocationChildren((prev) => ({
                ...prev,
                [id]: data ?? [],
            }))
        } catch {
            setLocationsError("Не удалось загрузить дочерние локации")
        } finally {
            setLocationLoading((prev) => {
                const next = new Set(prev)
                next.delete(id)
                return next
            })
        }
    }

    async function handleProfileSave() {
        if (!user) {
            return
        }
        setProfileSaving(true)
        setProfileError(null)
        try {
            const payload: Record<string, unknown> = {
                login: profileDraft.login.trim(),
                first_name: profileDraft.first_name.trim(),
                last_name: profileDraft.last_name.trim() || undefined,
                middle_name: profileDraft.middle_name.trim() || undefined,
                email: profileDraft.email.trim() || undefined,
                is_active: profileDraft.is_active,
            }
            if (profileDraft.password.trim()) {
                payload.password = profileDraft.password
            }
            await updateUser(user.id, payload)
            setProfileDraft((prev) => ({...prev, password: ""}))
        } catch {
            setProfileError("Не удалось обновить профиль")
        } finally {
            setProfileSaving(false)
        }
    }

    useEffect(() => {
        if (activeTab === "locations" && token) {
            loadBuildingLocations()
        }
    }, [activeTab, token, locationsLoaded])

    function getChildType(type: string) {
        if (type === "building") return "room"
        if (type === "room") return "cabinet"
        if (type === "cabinet") return "shelf"
        return null
    }

    function openAddLocation(type: string, parentId = "") {
        setLocationDraft({
            parentId,
            type,
            name: "",
            address: "",
            description: "",
            lockParent: !!parentId,
            lockType: !!type,
        })
        setLocationError(null)
        const parentType = getParentType(type)
        if (parentType) {
            ensureLocationsByType(parentType)
        }
        setIsLocationModalOpen(true)
    }

    function renderLocationNode(location: LocationEntity, level = 0) {
        const childType = getChildType(location.type)
        const isExpanded = expandedLocations.has(location.id)
        const children = locationChildren[location.id] ?? []
        const isLoading = locationLoading.has(location.id)

        return (
            <div key={location.id} className="location-node">
                <div className="location-row">
                    {childType ? (
                        <button
                            className={`icon-button ${
                                isExpanded ? "icon-rotated" : ""
                            }`}
                            type="button"
                            onClick={() => toggleLocation(location)}
                            aria-label={
                                isExpanded
                                    ? "Свернуть"
                                    : "Развернуть"
                            }
                        >
                            ▸
                        </button>
                    ) : (
                        <span className="icon-placeholder" />
                    )}
                    <span className="location-name">
                        {location.name}
                    </span>
                    {location.type === "building" && location.address && (
                        <span className="location-address">
                            {location.address}
                        </span>
                    )}
                    {childType && (
                        <button
                            className="icon-plus-button"
                            type="button"
                            onClick={() =>
                                openAddLocation(childType, location.id)
                            }
                            aria-label="Добавить дочернюю локацию"
                        >
                            +
                        </button>
                    )}
                </div>
                {isExpanded && (
                    <div className="location-children">
                        {isLoading && (
                            <span className="item-meta">Загрузка...</span>
                        )}
                        {!isLoading && children.length === 0 && (
                            <div className="location-empty">
                                Нет дочерних локаций
                            </div>
                        )}
                        {children.map((child) =>
                            renderLocationNode(child, level + 1)
                        )}
                    </div>
                )}
            </div>
        )
    }

    async function loadChildren(
        parentId: string,
        childType: string,
        reset = false
    ) {
        if (!parentId) {
            return
        }
        if (locationChildren[parentId] && !reset) {
            return
        }
        try {
            const data = await getLocationChildren(parentId, childType)
            setLocationChildren((prev) => ({
                ...prev,
                [parentId]: data ?? [],
            }))
        } catch {
            setLocationsError("Не удалось загрузить локации")
        }
    }

    return (
        <div className="app-shell">
            <header className="top-bar">
                <div>
                    <p className="eyebrow">электронная библиотека</p>
                    <h1>Поиск фонда и управление каталогом</h1>
                </div>
                <div className="user-block">
                    {user && token ? (
                        <>
                            <button
                                className="user-badge-button"
                                type="button"
                                onClick={() => setActiveTab("profile")}
                            >
                                {user.login || "Пользователь"}
                                {isAdmin ? " · админ" : ""}
                            </button>
                            <button
                                className="ghost-button"
                                type="button"
                                onClick={handleLogout}
                            >
                                Выйти
                            </button>
                        </>
                    ) : (
                        <button
                            className="user-badge-button"
                            type="button"
                            onClick={() => setIsLoginOpen(true)}
                        >
                            Войти
                        </button>
                    )}
                </div>
            </header>

            <nav className="tab-bar">
                {(
                    [
                        "books",
                        "works",
                        "authors",
                        "publishers",
                        "locations",
                    ] as TabKey[]
                ).map(
                    (tab) => (
                        <button
                            key={tab}
                            className={`tab-button ${
                                activeTab === tab ? "active" : ""
                            }`}
                            type="button"
                            onClick={() => setActiveTab(tab)}
                        >
                            {tab === "books" && "Книги"}
                            {tab === "works" && "Произведения"}
                            {tab === "authors" && "Авторы"}
                            {tab === "publishers" && "Издательства"}
                            {tab === "locations" && "Локации"}
                        </button>
                    )
                )}
            </nav>

            {authError && <p className="error-banner">{authError}</p>}

            {activeTab === "books" && (
                <section className="panel">
                    <div className="panel-header">
                        <div>
                            <h2>Поиск книг</h2>
                            <p className="results-caption">
                                Поиск по штрихкоду, названию, автору или
                                произведению.
                            </p>
                        </div>
                        {isAdmin && (
                            <div className="panel-actions">
                                <button
                                    className="primary-button"
                                    type="button"
                                    onClick={() => setIsBookModalOpen(true)}
                                >
                                    Добавить книгу
                                </button>
                            </div>
                        )}
                    </div>
                    <SearchBar
                        onSearch={handleBookSearch}
                        onReset={() => {
                            setBooks([])
                            setBooksQuery("")
                            setBooksError(null)
                        }}
                        isLoading={booksLoading}
                    />
                    {booksError && <p className="error-banner">{booksError}</p>}
                    {!booksLoading && books.length === 0 && booksQuery && (
                        <p className="status-line">Нет результатов</p>
                    )}
                    <div className="card-grid">
                        {books.map((book) => (
                            <article key={book.id} className="item-card">
                                <div className="item-header">
                                    <SafeImage
                                        src={getCoverUrl(book)}
                                        alt={book.title}
                                        className="cover-image"
                                    />
                                    <div>
                                        <h3>{book.title}</h3>
                                        <p className="item-meta">
                                            {book.publisher?.name ||
                                                "Без издательства"}
                                        </p>
                                    </div>
                                </div>
                                <div className="item-body">
                                    <p>
                                        <strong>Произведения:</strong>{" "}
                                            {book.works?.length
                                                ? book.works
                                                      .map((work) => work.title)
                                                      .join(", ")
                                                : "—"}
                                    </p>
                                    <p>
                                        <strong>Авторы:</strong>{" "}
                                        {book.works?.length
                                            ? Array.from(
                                                  new Set(
                                                      book.works.flatMap(
                                                          (work) =>
                                                              (work.authors ??
                                                                  []
                                                              ).map((author) =>
                                                                  getAuthorName(
                                                                      author
                                                                  )
                                                              )
                                                      )
                                                  )
                                              ).join(", ")
                                            : "—"}
                                    </p>
                                    {book.year && (
                                        <p>
                                            <strong>Год:</strong> {book.year}
                                        </p>
                                    )}
                                    {isAdmin && "location" in book && (
                                        <p>
                                            <strong>Локация:</strong>{" "}
                                            {formatLocation(
                                                (book as BookInternal).location
                                            )}
                                        </p>
                                    )}
                                </div>
                            </article>
                        ))}
                    </div>
                </section>
            )}

            {activeTab === "works" && (
                <section className="panel">
                    <div className="panel-header">
                        <div>
                            <h2>Произведения и книги</h2>
                            <p className="results-caption">
                                Выберите произведение и смотрите, в каких
                                книгах оно встречается.
                            </p>
                        </div>
                        {isAdmin && (
                            <button
                                className="primary-button"
                                type="button"
                                onClick={() => setIsWorkModalOpen(true)}
                            >
                                Добавить произведение
                            </button>
                        )}
                    </div>
                    <div className="split-layout">
                        <div>
                            <label className="field-label">
                                Поиск произведения
                            </label>
                            <input
                                className="text-input"
                                value={workQuery}
                                onChange={(event) =>
                                    setWorkQuery(event.target.value)
                                }
                                placeholder="Название произведения"
                            />
                            <div className="list">
                                {filteredWorks.map((work) => (
                                    <button
                                        key={work.id}
                                        className={`list-item ${
                                            selectedWork?.id === work.id
                                                ? "active"
                                                : ""
                                        }`}
                                        type="button"
                                        onClick={() => handleWorkSelect(work)}
                                    >
                                        <span>{work.title}</span>
                                        <span className="item-meta">
                                            {(() => {
                                                const names = (work.authors ?? [])
                                                    .map(getAuthorName)
                                                    .join(", ")
                                                const base = names || "Без автора"
                                                return work.year
                                                    ? `${base}, ${work.year}`
                                                    : base
                                            })()}
                                        </span>
                                    </button>
                                ))}
                            </div>
                        </div>
                        <div>
                            <h3 className="subheading">
                                Книги с этим произведением
                            </h3>
                            {workBooksLoading && (
                                <p className="status-line">Загрузка...</p>
                            )}
                            {!workBooksLoading && workBooks.length === 0 && (
                                <p className="status-line">Нет данных</p>
                            )}
                            <div className="stack">
                                {workBooks.map((book) => (
                                    <div key={book.id} className="mini-card">
                                        <span>{book.title}</span>
                                        {isAdmin && "location" in book && (
                                            <span className="item-meta">
                                                {formatLocation(
                                                    (book as BookInternal)
                                                        .location
                                                )}
                                            </span>
                                        )}
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>
                </section>
            )}

            {activeTab === "authors" && (
                <section className="panel">
                    <div className="panel-header">
                        <div>
                            <h2>Авторы и книги</h2>
                            <p className="results-caption">
                                Ищите автора и узнавайте, в каких книгах он
                                встречается.
                            </p>
                        </div>
                        {isAdmin && (
                            <button
                                className="primary-button"
                                type="button"
                                onClick={() => setIsAuthorModalOpen(true)}
                            >
                                Добавить автора
                            </button>
                        )}
                    </div>
                    <div className="split-layout">
                        <div>
                            <label className="field-label">Поиск автора</label>
                            <input
                                className="text-input"
                                value={authorQuery}
                                onChange={(event) =>
                                    setAuthorQuery(event.target.value)
                                }
                                placeholder="Фамилия или имя"
                            />
                            <div className="list">
                                    {filteredAuthors.map((author) => (
                                        <button
                                            key={author.id}
                                            className={`list-item ${
                                                selectedAuthor?.id === author.id
                                                    ? "active"
                                                    : ""
                                            }`}
                                            type="button"
                                            onClick={() => handleAuthorSelect(author)}
                                        >
                                            <span>{getAuthorName(author)}</span>
                                        </button>
                                    ))}
                            </div>
                        </div>
                        <div>
                            <p className="status-line">
                                Выберите автора слева, чтобы открыть карточку.
                            </p>
                        </div>
                    </div>
                </section>
            )}

            {activeTab === "publishers" && (
                <section className="panel">
                    <div className="panel-header">
                        <div>
                            <h2>Издательства</h2>
                            <p className="results-caption">
                                Просматривайте список издательств и их карточки.
                            </p>
                        </div>
                        {isAdmin && (
                            <button
                                className="primary-button"
                                type="button"
                                onClick={() => setIsPublisherModalOpen(true)}
                            >
                                Добавить издательство
                            </button>
                        )}
                    </div>
                    <label className="field-label">
                        Поиск издательства
                        <input
                            className="text-input"
                            value={publisherQuery}
                            onChange={(event) =>
                                setPublisherQuery(event.target.value)
                            }
                            placeholder="Название издательства"
                        />
                    </label>
                    <div className="card-grid">
                        {filteredPublishers.map((publisher) => (
                            <article key={publisher.id} className="item-card">
                                <div className="item-header">
                                    <SafeImage
                                        src={publisher.logo_url}
                                        alt={publisher.name}
                                        className="avatar"
                                    />
                                    <div>
                                        <h3>{publisher.name}</h3>
                                        {publisher.web_url && (
                                            <a
                                                className="item-meta"
                                                href={publisher.web_url}
                                                target="_blank"
                                                rel="noreferrer"
                                            >
                                                {publisher.web_url}
                                            </a>
                                        )}
                                    </div>
                                </div>
                            </article>
                        ))}
                    </div>
                </section>
            )}

            {activeTab === "locations" && (
                <section className="panel">
                    <div className="panel-header">
                        <div>
                            <h2>Локации</h2>
                            <p className="results-caption">
                                Управляйте иерархией расположения фонда.
                            </p>
                        </div>
                        {isAdmin && (
                            <button
                                className="primary-button"
                                type="button"
                                onClick={() => openAddLocation("building")}
                            >
                                Добавить новое здание
                            </button>
                        )}
                    </div>
                    {locationsError ? (
                        <p className="error-banner">{locationsError}</p>
                    ) : (
                        <div className="location-tree">
                            {!locationsLoaded && (
                                <p className="status-line">Загрузка...</p>
                            )}
                            {locationsLoaded && buildingLocations.length === 0 && (
                                <p className="status-line">
                                    Локаций пока нет.
                                </p>
                            )}
                            {buildingLocations.map((location) =>
                                renderLocationNode(location)
                            )}
                        </div>
                    )}
                </section>
            )}

            {activeTab === "profile" && (
                <section className="panel">
                    <div className="panel-header">
                        <div>
                            <h2>Личный кабинет</h2>
                            <p className="results-caption">
                                Управляйте своим аккаунтом и доступами.
                            </p>
                        </div>
                    </div>
                    {!token && (
                        <p className="status-line">
                            Чтобы войти в личный кабинет, авторизуйтесь.
                        </p>
                    )}
                    {token && user && (
                        <div className="profile-grid">
                            <div>
                                <h3>Профиль</h3>
                                <div className="form-grid">
                                    <label className="field-label">
                                        Логин
                                        <input
                                            className="text-input"
                                            value={profileDraft.login}
                                            onChange={(event) =>
                                                setProfileDraft((prev) => ({
                                                    ...prev,
                                                    login: event.target.value,
                                                }))
                                            }
                                            disabled={!isAdmin}
                                        />
                                    </label>
                                    <label className="field-label">
                                        Имя
                                        <input
                                            className="text-input"
                                            value={profileDraft.first_name}
                                            onChange={(event) =>
                                                setProfileDraft((prev) => ({
                                                    ...prev,
                                                    first_name:
                                                        event.target.value,
                                                }))
                                            }
                                            disabled={!isAdmin}
                                        />
                                    </label>
                                    <label className="field-label">
                                        Фамилия
                                        <input
                                            className="text-input"
                                            value={profileDraft.last_name}
                                            onChange={(event) =>
                                                setProfileDraft((prev) => ({
                                                    ...prev,
                                                    last_name:
                                                        event.target.value,
                                                }))
                                            }
                                            disabled={!isAdmin}
                                        />
                                    </label>
                                    <label className="field-label">
                                        Отчество
                                        <input
                                            className="text-input"
                                            value={profileDraft.middle_name}
                                            onChange={(event) =>
                                                setProfileDraft((prev) => ({
                                                    ...prev,
                                                    middle_name:
                                                        event.target.value,
                                                }))
                                            }
                                            disabled={!isAdmin}
                                        />
                                    </label>
                                    <label className="field-label">
                                        Email
                                        <input
                                            className="text-input"
                                            value={profileDraft.email}
                                            onChange={(event) =>
                                                setProfileDraft((prev) => ({
                                                    ...prev,
                                                    email: event.target.value,
                                                }))
                                            }
                                            disabled={!isAdmin}
                                        />
                                    </label>
                                    <label className="field-label">
                                        Новый пароль
                                        <input
                                            className="text-input"
                                            type="password"
                                            value={profileDraft.password}
                                            onChange={(event) =>
                                                setProfileDraft((prev) => ({
                                                    ...prev,
                                                    password:
                                                        event.target.value,
                                                }))
                                            }
                                            disabled={!isAdmin}
                                        />
                                    </label>
                                {!isAdmin && (
                                    <label className="field-label">
                                        <span>Активен</span>
                                        <input
                                            type="checkbox"
                                            checked={profileDraft.is_active}
                                            onChange={(event) =>
                                                setProfileDraft((prev) => ({
                                                    ...prev,
                                                    is_active:
                                                        event.target.checked,
                                                }))
                                            }
                                        />
                                    </label>
                                )}
                            </div>
                            {profileError && (
                                <p className="error-banner">
                                    {profileError}
                                </p>
                            )}
                            <div
                                className={`profile-actions ${
                                    !isAdmin ? "profile-actions-spaced" : ""
                                }`}
                            >
                                <button
                                    className="primary-button"
                                    type="button"
                                    onClick={handleProfileSave}
                                    disabled={profileSaving || !isAdmin}
                                >
                                    {profileSaving
                                        ? "Сохранение..."
                                        : "Сохранить изменения"}
                                </button>
                            </div>
                        </div>
                            <div>
                                <h3>Роли</h3>
                                <div className="tag-list">
                                    {user.roles.length ? (
                                        user.roles.map((role) => (
                                            <span key={role.code} className="tag">
                                                {role.name}
                                            </span>
                                        ))
                                    ) : (
                                        <span className="status-line">
                                            Нет данных
                                        </span>
                                    )}
                                </div>
                                {isAdmin && (
                                    <>
                                        <h3>Админ-инструменты</h3>
                                        <div className="stack">
                                            <button
                                                className="ghost-button"
                                                type="button"
                                                onClick={() =>
                                                    setIsPublisherModalOpen(true)
                                                }
                                            >
                                                Добавить издательство
                                            </button>
                                            <button
                                                className="ghost-button"
                                                type="button"
                                                onClick={() =>
                                                    openAddLocation("building")
                                                }
                                            >
                                                Добавить локацию
                                            </button>
                                        </div>
                                    </>
                                )}
                            </div>
                        </div>
                    )}
                </section>
            )}

            {isLoginOpen && (
                <div className="modal-backdrop">
                    <div className="modal">
                        <div className="modal-header">
                            <h3>Вход</h3>
                            <button
                                className="ghost-button"
                                type="button"
                                onClick={() => setIsLoginOpen(false)}
                            >
                                ✕
                            </button>
                        </div>
                        <form className="modal-body" onSubmit={handleLoginSubmit}>
                            <label className="field-label">
                                Логин
                                <input
                                    className="text-input"
                                    value={loginDraft.login}
                                    onChange={(event) =>
                                        setLoginDraft((prev) => ({
                                            ...prev,
                                            login: event.target.value,
                                        }))
                                    }
                                    required
                                />
                            </label>
                            <label className="field-label">
                                Пароль
                                <input
                                    className="text-input"
                                    type="password"
                                    value={loginDraft.password}
                                    onChange={(event) =>
                                        setLoginDraft((prev) => ({
                                            ...prev,
                                            password: event.target.value,
                                        }))
                                    }
                                    required
                                />
                            </label>
                            {authError && (
                                <p className="error-banner">{authError}</p>
                            )}
                            <div className="modal-footer">
                                <button className="primary-button" type="submit">
                                    Войти
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}

            {isBookModalOpen && (
                <div className="modal-backdrop">
                    <div className="modal modal-wide">
                        <div className="modal-header">
                            <h3>Новая книга</h3>
                            <button
                                className="ghost-button"
                                type="button"
                                onClick={() => setIsBookModalOpen(false)}
                            >
                                ✕
                            </button>
                        </div>
                        <div className="modal-body">
                            <div className="form-grid">
                                <label className="field-label">
                                    Название
                                    <input
                                        className="text-input"
                                        value={bookDraft.title}
                                        onChange={(event) =>
                                            setBookDraft((prev) => ({
                                                ...prev,
                                                title: event.target.value,
                                            }))
                                        }
                                    />
                                </label>
                                <label className="field-label">
                                    Издательство
                                    <div className="inline-actions">
                                        <select
                                            className={`text-input ${
                                                bookDraft.publisherId
                                                    ? ""
                                                    : "select-empty"
                                            }`}
                                            value={bookDraft.publisherId}
                                            onChange={(event) =>
                                                setBookDraft((prev) => ({
                                                    ...prev,
                                                    publisherId:
                                                        event.target.value,
                                                }))
                                            }
                                        >
                                            <option value="">Не выбрано</option>
                                            {publishers.map((publisher) => (
                                                <option
                                                    key={publisher.id}
                                                    value={publisher.id}
                                                >
                                                    {publisher.name}
                                                </option>
                                            ))}
                                        </select>
                                        <button
                                            className="ghost-button"
                                            type="button"
                                            onClick={() =>
                                                setIsPublisherModalOpen(true)
                                            }
                                            aria-label="Добавить издательство"
                                        >
                                            +
                                        </button>
                                    </div>
                                </label>
                                <label className="field-label">
                                    Год
                                    <input
                                        className="text-input"
                                        value={bookDraft.year}
                                        onChange={(event) =>
                                            setBookDraft((prev) => ({
                                                ...prev,
                                                year: event.target.value,
                                            }))
                                        }
                                    />
                                </label>
                                <label className="field-label">
                                    Локация
                                    <div className="location-selectors">
                                        <div className="inline-actions">
                                            <select
                                                className={`text-input ${
                                                    selectedBuildingId
                                                        ? ""
                                                        : "select-empty"
                                                }`}
                                                value={selectedBuildingId}
                                                onChange={(event) => {
                                                    const value = event.target.value
                                                    setSelectedBuildingId(value)
                                                    setSelectedRoomId("")
                                                    setSelectedCabinetId("")
                                                    setSelectedShelfId("")
                                                    setBookDraft((prev) => ({
                                                        ...prev,
                                                        locationId: "",
                                                    }))
                                                    if (value) {
                                                        loadChildren(value, "room", true)
                                                    }
                                                }}
                                            >
                                                <option value="">Не выбрано</option>
                                                {buildingLocations.map((location) => (
                                                    <option
                                                        key={location.id}
                                                        value={location.id}
                                                    >
                                                        {location.name}
                                                    </option>
                                                ))}
                                            </select>
                                            <button
                                                className="ghost-button"
                                                type="button"
                                                onClick={() => openAddLocation("building")}
                                                aria-label="Добавить здание"
                                            >
                                                +
                                            </button>
                                        </div>
                                        <div className="inline-actions">
                                            <select
                                                className={`text-input ${
                                                    selectedRoomId
                                                        ? ""
                                                        : "select-empty"
                                                }`}
                                                value={selectedRoomId}
                                                onChange={(event) => {
                                                    const value = event.target.value
                                                    setSelectedRoomId(value)
                                                    setSelectedCabinetId("")
                                                    setSelectedShelfId("")
                                                    setBookDraft((prev) => ({
                                                        ...prev,
                                                        locationId: "",
                                                    }))
                                                    if (value) {
                                                        loadChildren(value, "cabinet", true)
                                                    }
                                                }}
                                                disabled={!selectedBuildingId}
                                            >
                                                <option value="">Не выбрано</option>
                                                {roomLocations.map((location) => (
                                                    <option
                                                        key={location.id}
                                                        value={location.id}
                                                    >
                                                        {location.name}
                                                    </option>
                                                ))}
                                            </select>
                                            <button
                                                className="ghost-button"
                                                type="button"
                                                onClick={() =>
                                                    openAddLocation("room", selectedBuildingId)
                                                }
                                                aria-label="Добавить комнату"
                                                disabled={!selectedBuildingId}
                                            >
                                                +
                                            </button>
                                        </div>
                                        <div className="inline-actions">
                                            <select
                                                className={`text-input ${
                                                    selectedCabinetId
                                                        ? ""
                                                        : "select-empty"
                                                }`}
                                                value={selectedCabinetId}
                                                onChange={(event) => {
                                                    const value = event.target.value
                                                    setSelectedCabinetId(value)
                                                    setSelectedShelfId("")
                                                    setBookDraft((prev) => ({
                                                        ...prev,
                                                        locationId: "",
                                                    }))
                                                    if (value) {
                                                        loadChildren(value, "shelf", true)
                                                    }
                                                }}
                                                disabled={!selectedRoomId}
                                            >
                                                <option value="">Не выбрано</option>
                                                {cabinetLocations.map((location) => (
                                                    <option
                                                        key={location.id}
                                                        value={location.id}
                                                    >
                                                        {location.name}
                                                    </option>
                                                ))}
                                            </select>
                                            <button
                                                className="ghost-button"
                                                type="button"
                                                onClick={() =>
                                                    openAddLocation("cabinet", selectedRoomId)
                                                }
                                                aria-label="Добавить шкаф"
                                                disabled={!selectedRoomId}
                                            >
                                                +
                                            </button>
                                        </div>
                                        <div className="inline-actions">
                                            <select
                                                className={`text-input ${
                                                    selectedShelfId
                                                        ? ""
                                                        : "select-empty"
                                                }`}
                                                value={selectedShelfId}
                                                onChange={(event) => {
                                                    const value = event.target.value
                                                    setSelectedShelfId(value)
                                                    setBookDraft((prev) => ({
                                                        ...prev,
                                                        locationId: value,
                                                    }))
                                                }}
                                                disabled={!selectedCabinetId}
                                            >
                                                <option value="">Не выбрано</option>
                                                {shelfLocationsForSelection.map((location) => (
                                                    <option
                                                        key={location.id}
                                                        value={location.id}
                                                    >
                                                        {location.name}
                                                    </option>
                                                ))}
                                            </select>
                                            <button
                                                className="ghost-button"
                                                type="button"
                                                onClick={() =>
                                                    openAddLocation("shelf", selectedCabinetId)
                                                }
                                                aria-label="Добавить полку"
                                                disabled={!selectedCabinetId}
                                            >
                                                +
                                            </button>
                                        </div>
                                    </div>
                                </label>
                                <label className="field-label">
                                    Фабричный штрихкод
                                    <input
                                        className="text-input"
                                        value={bookDraft.factoryBarcode}
                                        onChange={(event) =>
                                            setBookDraft((prev) => ({
                                                ...prev,
                                                factoryBarcode:
                                                    event.target.value,
                                            }))
                                        }
                                    />
                                </label>
                                <label className="field-label">
                                    Обложка
                                    <input
                                        className="text-input"
                                        type="file"
                                        accept="image/*"
                                        onChange={(event) => {
                                            const file =
                                                event.target.files?.[0] ?? null
                                            if (!file) {
                                                setCoverPreview(null)
                                                return
                                            }
                                            const url = URL.createObjectURL(
                                                file
                                            )
                                            setCoverPreview(url)
                                        }}
                                    />
                                    <span className="item-meta">
                                        Пока используется только для предпросмотра.
                                    </span>
                                    {coverPreview && (
                                        <img
                                            className="cover-preview"
                                            src={coverPreview}
                                            alt="Предпросмотр обложки"
                                        />
                                    )}
                                </label>
                                <label className="field-label full">
                                    Описание
                                    <textarea
                                        className="text-area"
                                        value={bookDraft.description}
                                        onChange={(event) =>
                                            setBookDraft((prev) => ({
                                                ...prev,
                                                description:
                                                    event.target.value,
                                            }))
                                        }
                                    />
                                </label>
                            </div>
                            <div className="list-block">
                                <div className="list-block-header">
                                    <h4>Произведения</h4>
                                    <div className="inline-actions">
                                        <button
                                            className="ghost-button"
                                            type="button"
                                            onClick={() =>
                                                setIsWorksPickerOpen((prev) => !prev)
                                            }
                                        >
                                            Выбрать
                                        </button>
                                        <button
                                            className="ghost-button"
                                            type="button"
                                            onClick={() =>
                                                setIsWorkModalOpen(true)
                                            }
                                            aria-label="Добавить произведение"
                                        >
                                            +
                                        </button>
                                    </div>
                                </div>
                                {isWorksPickerOpen && (
                                    <div className="checkbox-grid">
                                        {works.map((work) => (
                                            <label
                                                key={work.id}
                                                className="checkbox-item"
                                            >
                                                <input
                                                    type="checkbox"
                                                    checked={bookDraft.workIds.includes(
                                                        work.id
                                                    )}
                                                    onChange={(event) => {
                                                        const checked =
                                                            event.target.checked
                                                        setBookDraft((prev) => ({
                                                            ...prev,
                                                            workIds: checked
                                                                ? [
                                                                      ...prev.workIds,
                                                                      work.id,
                                                                  ]
                                                                : prev.workIds.filter(
                                                                      (id) =>
                                                                          id !==
                                                                          work.id
                                                                  ),
                                                        }))
                                                    }}
                                                />
                                                <span>
                                                    {work.title}
                                                    <small className="item-meta">
                                                    {(() => {
                                                        const names = (work.authors ?? [])
                                                            .map(getAuthorName)
                                                            .join(", ")
                                                        const base =
                                                            names || "Без автора"
                                                        return work.year
                                                            ? `${base}, ${work.year}`
                                                            : base
                                                    })()}
                                                </small>
                                            </span>
                                        </label>
                                    ))}
                                    </div>
                                )}
                            </div>
                            {bookError && (
                                <p className="error-banner">{bookError}</p>
                            )}
                        </div>
                        <div className="modal-footer">
                            <button
                                className="primary-button"
                                type="button"
                                onClick={handleCreateBook}
                                disabled={bookSaving}
                            >
                                {bookSaving ? "Сохранение..." : "Добавить книгу"}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {isWorkModalOpen && (
                <div className="modal-backdrop">
                    <div className="modal">
                        <div className="modal-header">
                            <h3>Новое произведение</h3>
                            <button
                                className="ghost-button"
                                type="button"
                                onClick={() => setIsWorkModalOpen(false)}
                            >
                                ✕
                            </button>
                        </div>
                        <div className="modal-body">
                            <label className="field-label">
                                Название
                                <input
                                    className="text-input"
                                    value={workDraft.title}
                                    onChange={(event) =>
                                        setWorkDraft((prev) => ({
                                            ...prev,
                                            title: event.target.value,
                                        }))
                                    }
                                />
                            </label>
                            <label className="field-label">
                                Год
                                <input
                                    className="text-input"
                                    value={workDraft.year}
                                    onChange={(event) =>
                                        setWorkDraft((prev) => ({
                                            ...prev,
                                            year: event.target.value,
                                        }))
                                    }
                                />
                            </label>
                            <label className="field-label">
                                Описание
                                <textarea
                                    className="text-area"
                                    value={workDraft.description}
                                    onChange={(event) =>
                                        setWorkDraft((prev) => ({
                                            ...prev,
                                            description: event.target.value,
                                        }))
                                    }
                                />
                            </label>
                            <div className="list-block">
                                <div className="list-block-header">
                                    <h4>Авторы</h4>
                                    <button
                                        className="ghost-button"
                                        type="button"
                                        onClick={() =>
                                            setIsAuthorModalOpen(true)
                                        }
                                        aria-label="Добавить автора"
                                    >
                                        +
                                    </button>
                                </div>
                                <div className="checkbox-grid">
                                    {authors.map((author) => {
                                        const isSelected = workDraft.authorIds.includes(
                                            author.id
                                        )
                                        return (
                                            <div
                                                key={author.id}
                                                className="author-row"
                                            >
                                                <span>{getAuthorName(author)}</span>
                                                <button
                                                    className={`icon-plus-button ${
                                                        isSelected
                                                            ? "icon-minus"
                                                            : ""
                                                    }`}
                                                    type="button"
                                                    onClick={() => {
                                                        setWorkDraft((prev) => ({
                                                            ...prev,
                                                            authorIds: isSelected
                                                                ? prev.authorIds.filter(
                                                                      (id) =>
                                                                          id !==
                                                                          author.id
                                                                  )
                                                                : [
                                                                      ...prev.authorIds,
                                                                      author.id,
                                                                  ],
                                                        }))
                                                    }}
                                                    aria-label={
                                                        isSelected
                                                            ? "Убрать автора"
                                                            : "Добавить автора"
                                                    }
                                                >
                                                    {isSelected ? "−" : "+"}
                                                </button>
                                            </div>
                                        )
                                    })}
                                </div>
                            </div>
                            {workError && (
                                <p className="error-banner">{workError}</p>
                            )}
                        </div>
                        <div className="modal-footer">
                            <button
                                className="primary-button"
                                type="button"
                                onClick={handleCreateWork}
                                disabled={workSaving}
                            >
                                {workSaving
                                    ? "Сохранение..."
                                    : "Добавить произведение"}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {isAuthorModalOpen && (
                <div className="modal-backdrop">
                    <div className="modal">
                        <div className="modal-header">
                            <h3>Новый автор</h3>
                            <button
                                className="ghost-button"
                                type="button"
                                onClick={() => setIsAuthorModalOpen(false)}
                            >
                                ✕
                            </button>
                        </div>
                        <div className="modal-body">
                            <div className="form-grid">
                                <label className="field-label">
                                    Фамилия
                                    <input
                                        className="text-input"
                                        value={authorDraft.lastName}
                                        onChange={(event) =>
                                            setAuthorDraft((prev) => ({
                                                ...prev,
                                                lastName: event.target.value,
                                            }))
                                        }
                                    />
                                </label>
                                <label className="field-label">
                                    Имя
                                    <input
                                        className="text-input"
                                        value={authorDraft.firstName}
                                        onChange={(event) =>
                                            setAuthorDraft((prev) => ({
                                                ...prev,
                                                firstName: event.target.value,
                                            }))
                                        }
                                    />
                                </label>
                                <label className="field-label">
                                    Отчество
                                    <input
                                        className="text-input"
                                        value={authorDraft.middleName}
                                        onChange={(event) =>
                                            setAuthorDraft((prev) => ({
                                                ...prev,
                                                middleName: event.target.value,
                                            }))
                                        }
                                    />
                                </label>
                                <label className="field-label">
                                    Дата рождения
                                    <input
                                        className="text-input"
                                        type="date"
                                        value={authorDraft.birthDate}
                                        onChange={(event) =>
                                            setAuthorDraft((prev) => ({
                                                ...prev,
                                                birthDate: event.target.value,
                                            }))
                                        }
                                    />
                                </label>
                                <label className="field-label">
                                    Дата смерти
                                    <input
                                        className="text-input"
                                        type="date"
                                        value={authorDraft.deathDate}
                                        onChange={(event) =>
                                            setAuthorDraft((prev) => ({
                                                ...prev,
                                                deathDate: event.target.value,
                                            }))
                                        }
                                    />
                                </label>
                                <label className="field-label">
                                    Фото (URL)
                                    <input
                                        className="text-input"
                                        value={authorDraft.photoUrl}
                                        onChange={(event) =>
                                            setAuthorDraft((prev) => ({
                                                ...prev,
                                                photoUrl: event.target.value,
                                            }))
                                        }
                                    />
                                </label>
                                <label className="field-label full">
                                    Биография
                                    <textarea
                                        className="text-area"
                                        value={authorDraft.bio}
                                        onChange={(event) =>
                                            setAuthorDraft((prev) => ({
                                                ...prev,
                                                bio: event.target.value,
                                            }))
                                        }
                                    />
                                </label>
                            </div>
                            {authorError && (
                                <p className="error-banner">{authorError}</p>
                            )}
                        </div>
                        <div className="modal-footer">
                            <button
                                className="primary-button"
                                type="button"
                                onClick={handleCreateAuthor}
                                disabled={authorSaving}
                            >
                                {authorSaving ? "Сохранение..." : "Добавить автора"}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {isPublisherModalOpen && (
                <div className="modal-backdrop">
                    <div className="modal">
                        <div className="modal-header">
                            <h3>Новое издательство</h3>
                            <button
                                className="ghost-button"
                                type="button"
                                onClick={() => setIsPublisherModalOpen(false)}
                            >
                                ✕
                            </button>
                        </div>
                        <div className="modal-body">
                            <label className="field-label">
                                Название
                                <input
                                    className="text-input"
                                    value={publisherDraft.name}
                                    onChange={(event) =>
                                        setPublisherDraft((prev) => ({
                                            ...prev,
                                            name: event.target.value,
                                        }))
                                    }
                                />
                            </label>
                            <label className="field-label">
                                Логотип (URL)
                                <input
                                    className="text-input"
                                    value={publisherDraft.logoUrl}
                                    onChange={(event) =>
                                        setPublisherDraft((prev) => ({
                                            ...prev,
                                            logoUrl: event.target.value,
                                        }))
                                    }
                                />
                            </label>
                            <label className="field-label">
                                Сайт
                                <input
                                    className="text-input"
                                    value={publisherDraft.webUrl}
                                    onChange={(event) =>
                                        setPublisherDraft((prev) => ({
                                            ...prev,
                                            webUrl: event.target.value,
                                        }))
                                    }
                                />
                            </label>
                            {publisherError && (
                                <p className="error-banner">{publisherError}</p>
                            )}
                        </div>
                        <div className="modal-footer">
                            <button
                                className="primary-button"
                                type="button"
                                onClick={handleCreatePublisher}
                                disabled={publisherSaving}
                            >
                                {publisherSaving
                                    ? "Сохранение..."
                                    : "Добавить издательство"}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {isLocationModalOpen && (
                <div className="modal-backdrop">
                    <div className="modal">
                        <div className="modal-header">
                            <h3>Новая локация</h3>
                            <button
                                className="ghost-button"
                                type="button"
                                onClick={() => setIsLocationModalOpen(false)}
                            >
                                ✕
                            </button>
                        </div>
                        <div className="modal-body">
                            <label className="field-label">
                                Тип
                                <select
                                    className={`text-input ${
                                        locationDraft.type ? "" : "select-empty"
                                    }`}
                                    value={locationDraft.type}
                                    onChange={(event) => {
                                        const nextType = event.target.value
                                        setLocationDraft((prev) => ({
                                            ...prev,
                                            type: nextType,
                                            parentId: "",
                                            lockParent: false,
                                            lockType: prev.lockType,
                                        }))
                                        const parentType = getParentType(nextType)
                                        if (parentType) {
                                            ensureLocationsByType(parentType)
                                        }
                                    }}
                                    disabled={locationDraft.lockType}
                                >
                                    <option value="">Не выбрано</option>
                                    <option value="building">Здание</option>
                                    <option value="room">Комната</option>
                                    <option value="cabinet">Шкаф</option>
                                    <option value="shelf">Полка</option>
                                </select>
                            </label>
                            <label className="field-label">
                                Название
                                <input
                                    className="text-input"
                                    value={locationDraft.name}
                                    onChange={(event) =>
                                        setLocationDraft((prev) => ({
                                            ...prev,
                                            name: event.target.value,
                                        }))
                                    }
                                />
                            </label>
                            <label className="field-label">
                                Родитель
                                <select
                                    className={`text-input ${
                                        locationDraft.parentId ? "" : "select-empty"
                                    }`}
                                    value={locationDraft.parentId}
                                    onChange={(event) =>
                                        setLocationDraft((prev) => ({
                                            ...prev,
                                            parentId: event.target.value,
                                        }))
                                    }
                                    disabled={
                                        !getParentType(locationDraft.type) ||
                                        locationDraft.lockParent
                                    }
                                >
                                    <option value="">
                                        {getParentType(locationDraft.type)
                                            ? "Не выбрано"
                                            : "Не требуется"}
                                    </option>
                                    {(getParentType(locationDraft.type)
                                        ? locationByType[
                                              getParentType(locationDraft.type)!
                                          ] ?? []
                                        : []
                                    ).map((location) => (
                                        <option
                                            key={location.id}
                                            value={location.id}
                                        >
                                            {location.name}
                                        </option>
                                    ))}
                                </select>
                            </label>
                            <label className="field-label">
                                Адрес
                                <input
                                    className="text-input"
                                    value={locationDraft.address}
                                    onChange={(event) =>
                                        setLocationDraft((prev) => ({
                                            ...prev,
                                            address: event.target.value,
                                        }))
                                    }
                                />
                            </label>
                            <label className="field-label">
                                Описание
                                <textarea
                                    className="text-area"
                                    value={locationDraft.description}
                                    onChange={(event) =>
                                        setLocationDraft((prev) => ({
                                            ...prev,
                                            description: event.target.value,
                                        }))
                                    }
                                />
                            </label>
                            {locationError && (
                                <p className="error-banner">{locationError}</p>
                            )}
                        </div>
                        <div className="modal-footer">
                            <button
                                className="primary-button"
                                type="button"
                                onClick={handleCreateLocation}
                                disabled={locationSaving}
                            >
                                {locationSaving
                                    ? "Сохранение..."
                                    : "Добавить локацию"}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {isAuthorInfoOpen && (
                <div className="modal-backdrop">
                    <div className="modal">
                        <div className="modal-header">
                            <h3>Автор</h3>
                            <button
                                className="ghost-button"
                                type="button"
                                onClick={() => setIsAuthorInfoOpen(false)}
                            >
                                ✕
                            </button>
                        </div>
                        <div className="modal-body">
                            {selectedAuthor ? (
                                <div className="stack">
                                    <div className="author-card">
                                        <SafeImage
                                            src={selectedAuthor.photo_url}
                                            alt={getAuthorName(selectedAuthor)}
                                            className="avatar"
                                        />
                                        <div>
                                            <p className="author-name">
                                                {getAuthorName(selectedAuthor)}
                                            </p>
                                            {selectedAuthor.bio && (
                                                <p className="item-meta">
                                                    {selectedAuthor.bio}
                                                </p>
                                            )}
                                        </div>
                                    </div>
                                    <div>
                                        <h4 className="subheading">
                                            Книги, связанные с автором
                                        </h4>
                                        {authorBooksLoading && (
                                            <p className="status-line">
                                                Загрузка...
                                            </p>
                                        )}
                                        {!authorBooksLoading &&
                                            authorBooks.length === 0 && (
                                            <p className="status-line">
                                                Нет данных
                                            </p>
                                        )}
                                        <div className="stack">
                                            {authorBooks.map((book) => (
                                                <div
                                                    key={book.id}
                                                    className="mini-card"
                                                >
                                                    <span>{book.title}</span>
                                                    {isAdmin &&
                                                        "location" in book && (
                                                            <span className="item-meta">
                                                                {formatLocation(
                                                                    (
                                                                        book as BookInternal
                                                                    ).location
                                                                )}
                                                            </span>
                                                        )}
                                                </div>
                                            ))}
                                        </div>
                                    </div>
                                </div>
                            ) : (
                                <p className="status-line">Загрузка...</p>
                            )}
                        </div>
                    </div>
                </div>
            )}

            {isWorkInfoOpen && (
                <div className="modal-backdrop">
                    <div className="modal">
                        <div className="modal-header">
                            <h3>Произведение</h3>
                            <button
                                className="ghost-button"
                                type="button"
                                onClick={() => setIsWorkInfoOpen(false)}
                            >
                                ✕
                            </button>
                        </div>
                        <div className="modal-body">
                            {selectedWorkDetail ? (
                                <div className="stack">
                                    <h4>{selectedWorkDetail.title}</h4>
                                    <p className="item-meta">
                                        {(() => {
                                            const names = (selectedWorkDetail.authors ?? [])
                                                .map(getAuthorName)
                                                .join(", ")
                                            const base = names || "Без автора"
                                            return selectedWorkDetail.year
                                                ? `${base}, ${selectedWorkDetail.year}`
                                                : base
                                        })()}
                                    </p>
                                </div>
                            ) : (
                                <p className="status-line">Загрузка...</p>
                            )}
                        </div>
                        {isAdmin && (
                            <div className="modal-footer">
                                <button
                                    className="ghost-button"
                                    type="button"
                                    onClick={() => setIsWorkModalOpen(true)}
                                >
                                    Изменить
                                </button>
                            </div>
                        )}
                    </div>
                </div>
            )}
        </div>
    )
}

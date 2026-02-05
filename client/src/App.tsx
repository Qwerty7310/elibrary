import {useEffect, useMemo, useRef, useState} from "react"
import "./App.css"
import {loginUser} from "./api/auth"
import {
    createBook,
    searchBooksInternal,
    searchBooksPublic,
    updateBook,
} from "./api/books"
import {createAuthor, getAuthorByID, updateAuthor} from "./api/authors"
import {
    createPublisher,
    deletePublisher,
    getPublisherByID,
    updatePublisher,
} from "./api/publishers"
import {createWork, deleteWork, getWorkByID, updateWork} from "./api/works"
import {
    createLocation,
    getLocationChildren,
    getLocationsByType,
} from "./api/locations"
import {sendPrintTask} from "./api/print"
import {
    getAuthorsReference,
    getPublishersReference,
    getWorksReference,
} from "./api/reference"
import {API_URL, ApiError, setToken} from "./api/http"
import {getUserByID, updateUser} from "./api/users"
import {SafeImage} from "./components/SafeImage"
import {uploadImage} from "./api/images"
import type {
    Author,
    AuthorSummary,
    BookInternal,
    BookPublic,
    BookWorkInput,
    LocationEntity,
    Publisher,
    User,
    WorkDetailed,
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

type PrintQueueItem = {
    id: string
    title: string
    authors: string
    barcode: string
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
    const last = author.last_name?.trim()
    const first = author.first_name?.trim()
    const middle = author.middle_name?.trim()
    if (!last) {
        return [first, middle].filter(Boolean).join(" ")
    }
    if (!middle) {
        return [first, last].filter(Boolean).join(" ")
    }
    return [last, first, middle].filter(Boolean).join(" ")
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

function formatLocationShort(location?: BookInternal["location"]) {
    if (!location) {
        return "—"
    }
    const name =
        location.shelf_name ||
        location.cabinet_name ||
        location.room_name ||
        location.building_name
    const address = location.address ? `, ${location.address}` : ""
    return `${name}${address}`
}

function getBookAuthorsLine(book: BookPublic) {
    if (!book.works?.length) {
        return ""
    }
    return Array.from(
        new Set(
            book.works.flatMap((work) =>
                (work.authors ?? []).map((author) => getAuthorName(author))
            )
        )
    ).join(", ")
}

function getLocationPrintLine(location: LocationEntity) {
    const typeLabel = getLocationTypeLabel(location.type)
    if (location.type === "building" && location.address) {
        return `${typeLabel}: ${location.address}`
    }
    return typeLabel
}

function getLocationTypeLabel(type: string) {
    if (type === "building") return "Здание"
    if (type === "room") return "Комната"
    if (type === "cabinet") return "Шкаф"
    if (type === "shelf") return "Полка"
    return "Локация"
}

function formatDate(value?: string) {
    if (!value) {
        return ""
    }
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) {
        return value
    }
    const day = String(date.getUTCDate()).padStart(2, "0")
    const month = String(date.getUTCMonth() + 1).padStart(2, "0")
    const year = date.getUTCFullYear()
    return `${day}.${month}.${year}`
}

    function formatDateTime(value?: string) {
    if (!value) {
        return ""
    }
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) {
        return value
    }
    return date.toLocaleString("ru-RU", {
        day: "2-digit",
        month: "2-digit",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
    })
    }

    function formatLifeDates(author: Author) {
    const birth = formatDate(author.birth_date)
    const death = formatDate(author.death_date)
    if (!birth && !death) {
        return ""
    }

    return `${birth || "—"} - ${death || "—"}`
}

function getEntityImagePath(entity: "author" | "publisher" | "book", id: string) {
    return `${API_URL}/static/images/${entity}/${id}/photo.jpg`
}

function withCacheBust(url: string) {
    const stamp = Date.now()
    return url.includes("?") ? `${url}&v=${stamp}` : `${url}?v=${stamp}`
}

function truncateLabel(value: string, max = 32) {
    const trimmed = value.trim()
    if (trimmed.length <= max) {
        return trimmed
    }
    return `${trimmed.slice(0, Math.max(0, max - 1))}…`
}

function getCoverUrl(book: BookPublic) {
    const extra = book.extra ?? {}
    const cover = extra.cover_url
    if (typeof cover === "string" && cover.trim()) {
        return cover
    }
    return getEntityImagePath("book", book.id)
}

function isBookInternal(book: BookPublic | BookInternal): book is BookInternal {
    return "location" in book
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
    const [printError, setPrintError] = useState<string | null>(null)
    const [printQueue, setPrintQueue] = useState<PrintQueueItem[]>([])
    const [isPrintQueueOpen, setIsPrintQueueOpen] = useState(false)
    const [printQueueSending, setPrintQueueSending] = useState(false)
    const [selectedBook, setSelectedBook] = useState<BookPublic | null>(null)
    const [isBookInfoOpen, setIsBookInfoOpen] = useState(false)
    const [editingBookId, setEditingBookId] = useState<string | null>(null)

    const [workQuery, setWorkQuery] = useState("")
    const [selectedWork, setSelectedWork] = useState<WorkShort | null>(null)
    const [selectedWorkDetail, setSelectedWorkDetail] =
        useState<WorkDetailed | WorkShort | null>(null)
    const [isWorkInfoOpen, setIsWorkInfoOpen] = useState(false)
    const [workBooks, setWorkBooks] = useState<BookPublic[]>([])
    const [workBooksLoading, setWorkBooksLoading] = useState(false)

    const [authorQuery, setAuthorQuery] = useState("")
    const [selectedAuthor, setSelectedAuthor] = useState<Author | null>(null)
    const [selectedAuthorId, setSelectedAuthorId] = useState<string | null>(null)
    const [isAuthorInfoOpen, setIsAuthorInfoOpen] = useState(false)
    const [workAuthorSearch, setWorkAuthorSearch] = useState("")

    const [publisherQuery, setPublisherQuery] = useState("")
    const [selectedPublisher, setSelectedPublisher] = useState<Publisher | null>(
        null
    )
    const [isPublisherInfoOpen, setIsPublisherInfoOpen] = useState(false)
    const [publisherInfoError, setPublisherInfoError] = useState<string | null>(
        null
    )
    const [publisherInfoSaving, setPublisherInfoSaving] = useState(false)
    const [isPublisherEditOpen, setIsPublisherEditOpen] = useState(false)
    const [publisherEditDraft, setPublisherEditDraft] = useState({
        name: "",
        webUrl: "",
    })
    const [publisherEditLogoFile, setPublisherEditLogoFile] =
        useState<File | null>(null)
    const [publisherEditLogoPreview, setPublisherEditLogoPreview] =
        useState<string | null>(null)
    const [publisherEditSaving, setPublisherEditSaving] = useState(false)
    const [publisherEditError, setPublisherEditError] = useState<string | null>(
        null
    )

    const filteredBooks = useMemo(() => {
        if (!booksQuery.trim()) {
            return books
        }
        return books
    }, [books, booksQuery])

    const [isBookModalOpen, setIsBookModalOpen] = useState(false)
    const [bookDraft, setBookDraft] = useState<BookDraft>(emptyBookDraft)
    const [bookError, setBookError] = useState<string | null>(null)
    const [bookSaving, setBookSaving] = useState(false)
    const [coverFile, setCoverFile] = useState<File | null>(null)
    const [coverPreview, setCoverPreview] = useState<string | null>(null)
    const [coverFileName, setCoverFileName] = useState("")
    const [workSearch, setWorkSearch] = useState("")
    const [selectedBuildingId, setSelectedBuildingId] = useState("")
    const [selectedRoomId, setSelectedRoomId] = useState("")
    const [selectedCabinetId, setSelectedCabinetId] = useState("")
    const [selectedShelfId, setSelectedShelfId] = useState("")

    const [isWorkModalOpen, setIsWorkModalOpen] = useState(false)
    const [editingWorkId, setEditingWorkId] = useState<string | null>(null)
    const [workDraft, setWorkDraft] = useState<WorkDraft>(emptyWorkDraft)
    const [workError, setWorkError] = useState<string | null>(null)
    const [workSaving, setWorkSaving] = useState(false)

    const [isAuthorModalOpen, setIsAuthorModalOpen] = useState(false)
    const [authorDraft, setAuthorDraft] = useState<AuthorDraft>(emptyAuthorDraft)
    const [authorError, setAuthorError] = useState<string | null>(null)
    const [authorSaving, setAuthorSaving] = useState(false)
    const [authorPhotoFile, setAuthorPhotoFile] = useState<File | null>(null)
    const [authorPhotoPreview, setAuthorPhotoPreview] = useState<string | null>(
        null
    )
    const [authorInfoSaving] = useState(false)
    const [authorInfoError, setAuthorInfoError] = useState<string | null>(null)
    const [isAuthorEditOpen, setIsAuthorEditOpen] = useState(false)
    const [authorEditDraft, setAuthorEditDraft] = useState({
        lastName: "",
        firstName: "",
        middleName: "",
        birthDate: "",
        deathDate: "",
        bio: "",
    })
    const [authorEditPhotoFile, setAuthorEditPhotoFile] =
        useState<File | null>(null)
    const [authorEditPhotoPreview, setAuthorEditPhotoPreview] =
        useState<string | null>(null)
    const [authorEditSaving, setAuthorEditSaving] = useState(false)
    const [authorEditError, setAuthorEditError] = useState<string | null>(null)

    const [isPublisherModalOpen, setIsPublisherModalOpen] = useState(false)
    const [publisherDraft, setPublisherDraft] =
        useState<PublisherDraft>(emptyPublisherDraft)
    const [publisherError, setPublisherError] = useState<string | null>(null)
    const [publisherSaving, setPublisherSaving] = useState(false)
    const [publisherLogoFile, setPublisherLogoFile] = useState<File | null>(null)
    const [publisherLogoPreview, setPublisherLogoPreview] = useState<string | null>(
        null
    )

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
        if (
            !token ||
            activeTab !== "books" ||
            books.length > 0 ||
            booksQuery.trim()
        ) {
            return
        }
        void loadBooksAll()
    }, [token, activeTab, books.length, booksQuery])

    useEffect(() => {
        if (isBookModalOpen && token) {
            loadBuildingLocations()
            if (!editingBookId) {
                setSelectedBuildingId("")
                setSelectedRoomId("")
                setSelectedCabinetId("")
                setSelectedShelfId("")
                setBookDraft((prev) => ({...prev, locationId: ""}))
            }
            setWorkSearch("")
        }
    }, [isBookModalOpen, token, editingBookId])

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

    const selectedAuthorWorks = useMemo(() => {
        if (!selectedAuthorId) {
            return []
        }
        return works.filter((work) =>
            (work.authors ?? []).some((author) => author.id === selectedAuthorId)
        )
    }, [selectedAuthorId, works])

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
        if (!window.confirm("Вы действительно хотите выйти?")) {
            return
        }
        setAuthToken(null)
        setToken(null)
        setUser(null)
        setIsAdmin(false)
        localStorage.removeItem("login_name")
        setLoginName("")
        setActiveTab("books")
        setBooks([])
    }

    async function loadBooksAll() {
        if (!token) {
            setIsLoginOpen(true)
            return
        }
        setBooksLoading(true)
        setBooksError(null)
        try {
            const data = isAdmin
                ? await searchBooksInternal("")
                : await searchBooksPublic("")
            setBooks(data)
        } catch {
            setBooksError("Не удалось загрузить список книг")
        } finally {
            setBooksLoading(false)
        }
    }

    const bookSearchSeq = useRef(0)

    function handleBookQueryChange(value: string) {
        const normalized = value.replace(/\s+/g, " ").trimStart()
        setBooksQuery(normalized)

        const needle = normalized.trim()
        if (!needle) {
            void loadBooksAll()
            return
        }

        const requestId = ++bookSearchSeq.current
        setBooksLoading(true)
        setBooksError(null)
        ;(async () => {
            try {
                const data = isAdmin
                    ? await searchBooksInternal(needle)
                    : await searchBooksPublic(needle)
                if (requestId === bookSearchSeq.current) {
                    setBooks(data)
                }
            } catch {
                if (requestId === bookSearchSeq.current) {
                    setBooks([])
                    setBooksError("Не удалось найти книги")
                }
            } finally {
                if (requestId === bookSearchSeq.current) {
                    setBooksLoading(false)
                }
            }
        })()
    }

    async function handleWorkSelect(work: WorkShort, openInfo = true) {
        setSelectedWork(work)
        setSelectedWorkDetail(work)
        setIsWorkInfoOpen(openInfo)
        setWorkBooks([])
        setWorkBooksLoading(true)
        try {
            const details = await getWorkByID(work.id)
            setSelectedWorkDetail(details)
        } catch {
            // keep short details
        }
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

    function openWorkCreate() {
        setEditingWorkId(null)
        setWorkDraft(emptyWorkDraft)
        setWorkAuthorSearch("")
        setIsWorkModalOpen(true)
    }

    function openWorkEdit(work: WorkDetailed) {
        setEditingWorkId(work.id)
        setWorkDraft({
            title: work.title ?? "",
            description: work.description ?? "",
            year: work.year ? String(work.year) : "",
            authorIds: (work.authors ?? []).map((author) => author.id),
        })
        setWorkAuthorSearch("")
        setIsWorkModalOpen(true)
        setIsWorkInfoOpen(false)
    }

    function closeWorkModal() {
        setIsWorkModalOpen(false)
        setEditingWorkId(null)
        setWorkDraft(emptyWorkDraft)
        setWorkAuthorSearch("")
        setWorkError(null)
    }

    async function handleDeleteWork() {
        if (!selectedWorkDetail) {
            return
        }
        if (!window.confirm("Удалить произведение?")) {
            return
        }
        try {
            await deleteWork(selectedWorkDetail.id)
            setWorks((prev) =>
                prev.filter((work) => work.id !== selectedWorkDetail.id)
            )
            setSelectedWork(null)
            setSelectedWorkDetail(null)
            setWorkBooks([])
            setIsWorkInfoOpen(false)
        } catch {
            setWorkError("Не удалось удалить произведение")
        }
    }

    async function openWorkEditFromSelection() {
        if (!selectedWorkDetail) {
            return
        }
        try {
            const details = await getWorkByID(selectedWorkDetail.id)
            openWorkEdit(details)
        } catch {
            openWorkEdit({
                id: selectedWorkDetail.id,
                title: selectedWorkDetail.title,
                description: "",
                year: selectedWorkDetail.year,
                authors: selectedWorkDetail.authors ?? [],
            })
        }
    }

    function handleAuthorSelect(author: AuthorSummary) {
        setSelectedAuthorId(author.id)
    }

    async function handleAuthorInfoOpen(author: AuthorSummary) {
        setSelectedAuthor(null)
        setSelectedAuthorId(author.id)
        setIsAuthorInfoOpen(true)
        setAuthorInfoError(null)
        try {
            const details = await getAuthorByID(author.id)
            setSelectedAuthor(details)
        } catch {
            setSelectedAuthor(null)
        }
    }

    function openAuthorEdit(author: Author) {
        setIsAuthorInfoOpen(false)
        setAuthorEditDraft({
            lastName: author.last_name ?? "",
            firstName: author.first_name ?? "",
            middleName: author.middle_name ?? "",
            birthDate: author.birth_date ? author.birth_date.slice(0, 10) : "",
            deathDate: author.death_date ? author.death_date.slice(0, 10) : "",
            bio: author.bio ?? "",
        })
        setAuthorEditPhotoFile(null)
        if (authorEditPhotoPreview) {
            URL.revokeObjectURL(authorEditPhotoPreview)
        }
        setAuthorEditPhotoPreview(null)
        setAuthorEditError(null)
        setIsAuthorEditOpen(true)
    }

    async function handlePublisherInfoOpen(publisher: Publisher) {
        setSelectedPublisher(null)
        setIsPublisherInfoOpen(true)
        setPublisherInfoError(null)
        try {
            const details = await getPublisherByID(publisher.id)
            setSelectedPublisher(details)
        } catch {
            setSelectedPublisher(null)
        }
    }

    function openPublisherEdit(publisher: Publisher) {
        setPublisherEditDraft({
            name: publisher.name ?? "",
            webUrl: publisher.web_url ?? "",
        })
        setPublisherEditLogoFile(null)
        if (publisherEditLogoPreview) {
            URL.revokeObjectURL(publisherEditLogoPreview)
        }
        setPublisherEditLogoPreview(null)
        setPublisherEditError(null)
        setIsPublisherEditOpen(true)
    }

    async function openBookInfo(book: BookPublic) {
        setSelectedBook(book)
        setIsBookInfoOpen(true)
        if (!isAdmin || !token) {
            return
        }
        try {
            const internalBooks = await searchBooksInternal("")
            const found = internalBooks.find((item) => item.id === book.id)
            if (found) {
                setSelectedBook(found)
            }
        } catch {
            // keep public data if internal fetch fails
        }
    }

    function applyBookLocation(location?: BookInternal["location"]) {
        const buildingId = location?.building_id ?? ""
        const roomId = location?.room_id ?? ""
        const cabinetId = location?.cabinet_id ?? ""
        const shelfId = location?.shelf_id ?? ""
        const locationId = shelfId || cabinetId || roomId || buildingId || ""
        setSelectedBuildingId(buildingId)
        setSelectedRoomId(roomId)
        setSelectedCabinetId(cabinetId)
        setSelectedShelfId(shelfId)
        if (buildingId) {
            void loadChildren(buildingId, "room", true)
        }
        if (roomId) {
            void loadChildren(roomId, "cabinet", true)
        }
        if (cabinetId) {
            void loadChildren(cabinetId, "shelf", true)
        }
        return locationId
    }

    async function openBookEdit(book: BookPublic) {
        if (!isAdmin) {
            return
        }
        let details: BookInternal | BookPublic = book
        let location: BookInternal["location"] | undefined =
            isBookInternal(book) && book.location ? book.location : undefined
        if (!location && token) {
            try {
                const internalBooks = await searchBooksInternal("")
                const found = internalBooks.find((item) => item.id === book.id)
                if (found) {
                    details = found
                    location = found.location
                }
            } catch {
                // fallback to provided book data
            }
        }
        setSelectedBook(details)
        const locationId = applyBookLocation(location)
        setEditingBookId(book.id)
        setBookDraft({
            title: book.title ?? "",
            publisherId: book.publisher?.id ?? "",
            year: book.year ? String(book.year) : "",
            description: book.description ?? "",
            locationId,
            factoryBarcode: book.factory_barcode ?? "",
            workIds: (book.works ?? []).map((work) => work.id),
        })
        setCoverFile(null)
        if (coverPreview) {
            URL.revokeObjectURL(coverPreview)
        }
        setCoverPreview(null)
        setCoverFileName("")
        setWorkSearch("")
        setIsBookModalOpen(true)
        setIsBookInfoOpen(false)
    }

    async function handleCreateBook() {
        const isEditing = Boolean(editingBookId)
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
        const bookPayload = {
            title: bookDraft.title.trim(),
            publisher_id: bookDraft.publisherId || undefined,
            year: bookDraft.year ? Number(bookDraft.year) : undefined,
            description: bookDraft.description.trim() || undefined,
            location_id: bookDraft.locationId || undefined,
            factory_barcode: bookDraft.factoryBarcode.trim() || undefined,
        }
        try {
            if (isEditing && editingBookId) {
                await updateBook(editingBookId, {
                    ...bookPayload,
                    works: worksPayload,
                })
                if (coverFile) {
                    const coverUrl = await uploadImage(
                        "book",
                        editingBookId,
                        coverFile
                    )
                    await updateBook(editingBookId, {extra: {cover_url: coverUrl}})
                }
            } else {
                const created = await createBook({
                    book: {
                        ...bookPayload,
                        extra: undefined,
                    },
                    works: worksPayload,
                })
                if (!created.works || created.works.length === 0) {
                    created.works = works.filter((work) =>
                        bookDraft.workIds.includes(work.id)
                    )
                }
                if (coverFile) {
                    try {
                        const coverUrl = await uploadImage(
                            "book",
                            created.id,
                            coverFile
                        )
                        const extra = {
                            ...(created.extra ?? {}),
                            cover_url: coverUrl,
                        }
                        await updateBook(created.id, {extra})
                    } catch {
                        setBookError(
                            "Книга создана, но не удалось загрузить обложку"
                        )
                    }
                }
                addBookToPrintQueue(created)
            }
            setBookDraft(emptyBookDraft)
            setCoverFile(null)
            if (coverPreview) {
                URL.revokeObjectURL(coverPreview)
            }
            setCoverPreview(null)
            setCoverFileName("")
            setIsBookModalOpen(false)
            setEditingBookId(null)
            await loadBooksAll()
        } catch {
            setBookError(
                isEditing ? "Не удалось обновить книгу" : "Не удалось создать книгу"
            )
        } finally {
            setBookSaving(false)
        }
    }

    async function handleCreateWork() {
        const isEditing = Boolean(editingWorkId)
        if (!workDraft.title.trim()) {
            setWorkError("Название произведения обязательно")
            return
        }
        setWorkSaving(true)
        setWorkError(null)
        try {
            if (isEditing && editingWorkId) {
                await updateWork(editingWorkId, {
                    title: workDraft.title.trim(),
                    description: workDraft.description.trim(),
                    year: workDraft.year ? Number(workDraft.year) : undefined,
                    authors: workDraft.authorIds,
                })
                const selectedAuthors = authors.filter((author) =>
                    workDraft.authorIds.includes(author.id)
                )
                const year = workDraft.year ? Number(workDraft.year) : undefined
                setWorks((prev) =>
                    prev.map((work) =>
                        work.id === editingWorkId
                            ? {
                                  ...work,
                                  title: workDraft.title.trim(),
                                  authors: selectedAuthors,
                                  year,
                              }
                            : work
                    )
                )
                setSelectedWorkDetail((prev) =>
                    prev && prev.id === editingWorkId
                        ? {
                              ...prev,
                              title: workDraft.title.trim(),
                              description: workDraft.description.trim(),
                              year,
                              authors: selectedAuthors,
                          }
                        : prev
                )
            } else {
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
                if (isBookModalOpen) {
                    setBookDraft((prev) => ({
                        ...prev,
                        workIds: prev.workIds.includes(created.id)
                            ? prev.workIds
                            : [...prev.workIds, created.id],
                    }))
                }
            }
            setWorkDraft(emptyWorkDraft)
            setEditingWorkId(null)
            setIsWorkModalOpen(false)
        } catch {
            setWorkError(
                isEditing
                    ? "Не удалось обновить произведение"
                    : "Не удалось создать произведение"
            )
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
                photo_url: authorPhotoFile
                    ? undefined
                    : authorDraft.photoUrl.trim() || undefined,
            })
            if (authorPhotoFile) {
                try {
                    const photoUrl = await uploadImage(
                        "author",
                        created.id,
                        authorPhotoFile
                    )
                    await updateAuthor(created.id, {photo_url: photoUrl})
                    created.photo_url = withCacheBust(photoUrl)
                } catch {
                    setAuthorError(
                        "Автор создан, но не удалось загрузить фото"
                    )
                }
            }
            const summary = {
                id: created.id,
                last_name: created.last_name,
                first_name: created.first_name,
                middle_name: created.middle_name,
            }
            setAuthors((prev) => [summary, ...prev])
            setAuthorDraft(emptyAuthorDraft)
            setAuthorPhotoFile(null)
            if (authorPhotoPreview) {
                URL.revokeObjectURL(authorPhotoPreview)
            }
            setAuthorPhotoPreview(null)
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
                logo_url: publisherLogoFile
                    ? undefined
                    : publisherDraft.logoUrl.trim() || undefined,
                web_url: publisherDraft.webUrl.trim() || undefined,
            })
            let updatedPublisher = created
            if (publisherLogoFile) {
                try {
                    const logoUrl = await uploadImage(
                        "publisher",
                        created.id,
                        publisherLogoFile
                    )
                    await updatePublisher(created.id, {logo_url: logoUrl})
                    updatedPublisher = {
                        ...created,
                        logo_url: withCacheBust(logoUrl),
                    }
                } catch {
                    setPublisherError(
                        "Издательство создано, но не удалось загрузить логотип"
                    )
                }
            }
            setPublishers((prev) => [updatedPublisher, ...prev])
            setPublisherDraft(emptyPublisherDraft)
            setPublisherLogoFile(null)
            if (publisherLogoPreview) {
                URL.revokeObjectURL(publisherLogoPreview)
            }
            setPublisherLogoPreview(null)
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

    async function handleAuthorEditSave() {
        if (!selectedAuthor) {
            return
        }
        if (!authorEditDraft.lastName.trim()) {
            setAuthorEditError("Фамилия обязательна")
            return
        }
        setAuthorEditSaving(true)
        setAuthorEditError(null)
        try {
            await updateAuthor(selectedAuthor.id, {
                last_name: authorEditDraft.lastName.trim(),
                first_name: authorEditDraft.firstName.trim() || undefined,
                middle_name: authorEditDraft.middleName.trim() || undefined,
                birth_date: authorEditDraft.birthDate || undefined,
                death_date: authorEditDraft.deathDate || undefined,
                bio: authorEditDraft.bio.trim() || undefined,
            })
            let photoUrl = selectedAuthor.photo_url
            if (authorEditPhotoFile) {
                const uploadedUrl = await uploadImage(
                    "author",
                    selectedAuthor.id,
                    authorEditPhotoFile
                )
                await updateAuthor(selectedAuthor.id, {photo_url: uploadedUrl})
                photoUrl = withCacheBust(uploadedUrl)
            }
            const updated = {
                ...selectedAuthor,
                last_name: authorEditDraft.lastName.trim(),
                first_name: authorEditDraft.firstName.trim() || undefined,
                middle_name: authorEditDraft.middleName.trim() || undefined,
                birth_date: authorEditDraft.birthDate || undefined,
                death_date: authorEditDraft.deathDate || undefined,
                bio: authorEditDraft.bio.trim() || undefined,
                photo_url: photoUrl,
            }
            setSelectedAuthor(updated)
            setAuthors((prev) =>
                prev.map((author) =>
                    author.id === selectedAuthor.id
                        ? {
                              ...author,
                              last_name: updated.last_name,
                              first_name: updated.first_name,
                              middle_name: updated.middle_name,
                              photo_url: updated.photo_url,
                          }
                        : author
                )
            )
            setIsAuthorEditOpen(false)
        } catch {
            setAuthorEditError("Не удалось обновить автора")
        } finally {
            setAuthorEditSaving(false)
        }
    }


    async function handlePublisherEditSave() {
        if (!selectedPublisher) {
            return
        }
        if (!publisherEditDraft.name.trim()) {
            setPublisherEditError("Название издательства обязательно")
            return
        }
        setPublisherEditSaving(true)
        setPublisherEditError(null)
        try {
            const payload = {
                name: publisherEditDraft.name.trim(),
                web_url: publisherEditDraft.webUrl.trim() || undefined,
            }
            await updatePublisher(selectedPublisher.id, payload)
            let logoUrl = selectedPublisher.logo_url
            if (publisherEditLogoFile) {
                const uploadedUrl = await uploadImage(
                    "publisher",
                    selectedPublisher.id,
                    publisherEditLogoFile
                )
                await updatePublisher(selectedPublisher.id, {logo_url: uploadedUrl})
                logoUrl = withCacheBust(uploadedUrl)
            }
            const updated = {
                ...selectedPublisher,
                name: payload.name,
                web_url: payload.web_url,
                logo_url: logoUrl,
            }
            setSelectedPublisher(updated)
            setPublishers((prev) =>
                prev.map((publisher) =>
                    publisher.id === selectedPublisher.id
                        ? {...publisher, ...updated}
                        : publisher
                )
            )
            setIsPublisherEditOpen(false)
        } catch {
            setPublisherEditError("Не удалось обновить издательство")
        } finally {
            setPublisherEditSaving(false)
        }
    }

    async function handlePublisherDelete() {
        if (!selectedPublisher) {
            return
        }
        setPublisherInfoSaving(true)
        setPublisherInfoError(null)
        try {
            await deletePublisher(selectedPublisher.id)
            setPublishers((prev) =>
                prev.filter((publisher) => publisher.id !== selectedPublisher.id)
            )
            setSelectedPublisher(null)
            setIsPublisherInfoOpen(false)
        } catch {
            setPublisherInfoError("Не удалось удалить издательство")
        } finally {
            setPublisherInfoSaving(false)
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
                address:
                    locationDraft.type === "building"
                        ? locationDraft.address.trim() || undefined
                        : undefined,
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

    async function handlePrintTask(payload: {
        str1: string
        str2: string
        barcode: string
    }) {
        if (!payload.barcode.trim()) {
            setPrintError("Штрихкод отсутствует")
            return
        }
        try {
            setPrintError(null)
            await sendPrintTask(payload)
        } catch (err) {
            if (err instanceof ApiError) {
                setPrintError(err.message)
            } else {
                setPrintError("Не удалось отправить на печать")
            }
        }
    }

    function addBookToPrintQueue(book: BookPublic) {
        const barcode = (book.barcode ?? "").trim()
        if (!barcode) {
            setPrintError("Штрихкод отсутствует")
            return
        }
        const authorsLine = getBookAuthorsLine(book)
        setPrintError(null)
        setPrintQueue((prev) => {
            if (prev.some((item) => item.barcode === barcode)) {
                return prev
            }
            return [
                ...prev,
                {
                    id: book.id,
                    title: book.title,
                    authors: authorsLine,
                    barcode,
                },
            ]
        })
        setIsPrintQueueOpen(true)
    }

    async function sendPrintQueue() {
        if (printQueue.length === 0) {
            setPrintError("Очередь печати пуста")
            return
        }
        setPrintQueueSending(true)
        setPrintError(null)
        try {
            for (const item of printQueue) {
                await sendPrintTask({
                    str1: item.authors || "—",
                    str2: item.title,
                    barcode: item.barcode,
                })
            }
            setPrintQueue([])
            setIsPrintQueueOpen(false)
        } catch (err) {
            if (err instanceof ApiError) {
                setPrintError(err.message)
            } else {
                setPrintError("Не удалось отправить очередь на печать")
            }
        } finally {
            setPrintQueueSending(false)
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
                    <div className="location-row-left">
                        {childType ? (
                            <button
                                className={`icon-button ${
                                    isExpanded ? "icon-rotated" : ""
                                }`}
                                type="button"
                                onClick={() => toggleLocation(location)}
                                aria-label={
                                    isExpanded ? "Свернуть" : "Развернуть"
                                }
                            >
                                ▸
                            </button>
                        ) : (
                            <span className="icon-placeholder" />
                        )}
                    </div>
                    <div className="location-row-main">
                        <div className="location-title">
                            <span className="location-name">
                                {location.name}
                            </span>
                            <span
                                className={`location-type-badge type-${location.type}`}
                            >
                                {getLocationTypeLabel(location.type)}
                            </span>
                        </div>
                        <div className="location-meta">
                            {location.type === "building" && location.address && (
                                <span className="location-address">
                                    {location.address}
                                </span>
                            )}
                            <span className="location-barcode">
                                {location.barcode}
                            </span>
                        </div>
                    </div>
                    <div className="location-row-actions">
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
                        {isAdmin && (
                            <button
                                className="ghost-button"
                                type="button"
                                onClick={() =>
                                    handlePrintTask({
                                        str1: location.name,
                                        str2: getLocationPrintLine(location),
                                        barcode: location.barcode,
                                    })
                                }
                            >
                                Печать
                            </button>
                        )}
                    </div>
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
                    <div className="print-queue">
                        <button
                            className="print-queue-button"
                            type="button"
                            onClick={() =>
                                setIsPrintQueueOpen((prev) => !prev)
                            }
                            aria-label="Открыть очередь печати"
                        >
                            Очередь
                            <span className="print-queue-count">
                                {printQueue.length}
                            </span>
                        </button>
                        {isPrintQueueOpen && (
                            <div className="print-queue-panel">
                                <div className="print-queue-header">
                                    <strong>Очередь печати</strong>
                                    <button
                                        className="icon-button close-button"
                                        type="button"
                                        onClick={() =>
                                            setIsPrintQueueOpen(false)
                                        }
                                    >
                                        ✕
                                    </button>
                                </div>
                                <div className="print-queue-list">
                                    {printQueue.length === 0 && (
                                        <p className="status-line">
                                            Очередь пуста
                                        </p>
                                    )}
                                    {printQueue.map((item) => (
                                        <div
                                            key={item.barcode}
                                            className="print-queue-item"
                                        >
                                            <div>
                                                <div className="print-queue-title">
                                                    {item.title}
                                                </div>
                                                <div className="item-meta">
                                                    {item.authors || "—"}
                                                </div>
                                                <div className="item-meta">
                                                    {item.barcode}
                                                </div>
                                            </div>
                                            <button
                                                className="ghost-button"
                                                type="button"
                                                onClick={() =>
                                                    setPrintQueue((prev) =>
                                                        prev.filter(
                                                            (entry) =>
                                                                entry.barcode !==
                                                                item.barcode
                                                        )
                                                    )
                                                }
                                            >
                                                Убрать
                                            </button>
                                        </div>
                                    ))}
                                </div>
                                <div className="print-queue-actions">
                                    <button
                                        className="ghost-button"
                                        type="button"
                                        onClick={() => setPrintQueue([])}
                                        disabled={printQueue.length === 0}
                                    >
                                        Очистить
                                    </button>
                                    <button
                                        className="primary-button"
                                        type="button"
                                        onClick={sendPrintQueue}
                                        disabled={
                                            printQueueSending ||
                                            printQueue.length === 0
                                        }
                                    >
                                        {printQueueSending
                                            ? "Печать..."
                                            : "Печать"}
                                    </button>
                                </div>
                            </div>
                        )}
                    </div>
                    {user && token ? (
                        <>
                            <button
                                className="user-badge-button"
                                type="button"
                                onClick={() => setActiveTab("profile")}
                            >
                                {user.login || "Пользователь"}
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
                                    onClick={() => {
                                        setSelectedBook(null)
                                        setEditingBookId(null)
                                        setBookDraft(emptyBookDraft)
                                        setCoverFile(null)
                                        if (coverPreview) {
                                            URL.revokeObjectURL(coverPreview)
                                        }
                                        setCoverPreview(null)
                                        setCoverFileName("")
                                        setWorkSearch("")
                                        setIsBookModalOpen(true)
                                    }}
                                >
                                    Добавить книгу
                                </button>
                            </div>
                        )}
                    </div>
                    <label className="field-label">
                        Поиск книг
                        <input
                            className="text-input"
                            value={booksQuery}
                            onChange={(event) =>
                                handleBookQueryChange(event.target.value)
                            }
                            placeholder="Название, автор или штрихкод"
                        />
                    </label>
                    <div className="inline-actions">
                        <button
                            className="ghost-button"
                            type="button"
                            onClick={() => {
                                setBooksQuery("")
                                setBooksError(null)
                                void loadBooksAll()
                            }}
                            disabled={booksLoading || booksQuery.length === 0}
                        >
                            Сброс
                        </button>
                    </div>
                    {booksError && <p className="error-banner">{booksError}</p>}
                    {printError && <p className="error-banner">{printError}</p>}
                    {!booksLoading &&
                        filteredBooks.length === 0 &&
                        booksQuery && (
                        <p className="status-line">Нет результатов</p>
                    )}
                    <div className="card-grid books-grid">
                        {filteredBooks.map((book) => (
                            <article
                                key={book.id}
                                className="item-card"
                                onClick={() => openBookInfo(book)}
                                role="button"
                                tabIndex={0}
                            >
                                <div className="item-header">
                                    <SafeImage
                                        src={getCoverUrl(book)}
                                        alt={book.title}
                                        className="cover-image"
                                    />
                                    <div className="book-card-content">
                                        <h3>{book.title}</h3>
                                        <div className="book-card-meta">
                                            <p>
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
                                            <p className="item-meta">
                                                {book.publisher?.name ||
                                                    "Без издательства"}
                                            </p>
                                            {book.year && <p>{book.year}</p>}
                                        </div>
                                    </div>
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
                                onClick={openWorkCreate}
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
                                    <div
                                        key={work.id}
                                        className={`list-item ${
                                            selectedWork?.id === work.id
                                                ? "active"
                                                : ""
                                        }`}
                                    >
                                        <button
                                            className="list-item-button"
                                            type="button"
                                            onClick={() =>
                                                handleWorkSelect(work, false)
                                            }
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
                                        <button
                                            className="icon-button info-button"
                                            type="button"
                                            aria-label="Открыть информацию о произведении"
                                            onClick={() => handleWorkSelect(work, true)}
                                        >
                                            <em>i</em>
                                        </button>
                                    </div>
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
                                                {formatLocationShort(
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
                            <h2>Авторы и произведения</h2>
                            <p className="results-caption">
                                Ищите автора и узнавайте, в каких произведениях он
                                участвует.
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
                                    <div
                                        key={author.id}
                                        className={`list-item ${
                                            selectedAuthorId === author.id
                                                ? "active"
                                                : ""
                                        }`}
                                    >
                                        <SafeImage
                                            src={
                                                author.photo_url ||
                                                getEntityImagePath(
                                                    "author",
                                                    author.id
                                                )
                                            }
                                            alt={getAuthorName(author)}
                                            className="avatar list-avatar"
                                        />
                                        <button
                                            className="list-item-button"
                                            type="button"
                                            onClick={() =>
                                                handleAuthorSelect(author)
                                            }
                                        >
                                            <span>{getAuthorName(author)}</span>
                                        </button>
                                        <button
                                            className="icon-button info-button"
                                            type="button"
                                            aria-label="Открыть информацию об авторе"
                                            onClick={() =>
                                                handleAuthorInfoOpen(author)
                                            }
                                        >
                                            <em>i</em>
                                        </button>
                                    </div>
                                ))}
                            </div>
                        </div>
                        <div>
                            <h3 className="subheading">
                                Произведения с этим автором
                            </h3>
                            {!selectedAuthorId && (
                                <p className="status-line">
                                    Выберите автора слева.
                                </p>
                            )}
                            {selectedAuthorId &&
                                selectedAuthorWorks.length === 0 && (
                                <p className="status-line">Нет данных</p>
                            )}
                            <div className="stack">
                                {selectedAuthorWorks.map((work) => (
                                    <button
                                        key={work.id}
                                        className="list-item list-item-selectable"
                                        type="button"
                                        onClick={async () => {
                                            setActiveTab("works")
                                            await handleWorkSelect(work, false)
                                        }}
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
                            <article
                                key={publisher.id}
                                className="item-card publisher-card"
                            >
                                <div className="item-header">
                                    <SafeImage
                                        src={
                                            publisher.logo_url ||
                                            getEntityImagePath(
                                                "publisher",
                                                publisher.id
                                            )
                                        }
                                        alt={publisher.name}
                                        className="avatar"
                                    />
                                    <div>
                                        <h3>{publisher.name}</h3>
                                    </div>
                                    <button
                                        className="icon-button info-button"
                                        type="button"
                                        aria-label="Открыть информацию об издательстве"
                                        onClick={() =>
                                            handlePublisherInfoOpen(publisher)
                                        }
                                    >
                                        <em>i</em>
                                    </button>
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
                    {locationsError || printError ? (
                        <p className="error-banner">
                            {locationsError ?? printError}
                        </p>
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
                                className="icon-button close-button"
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

            {isBookModalOpen && isAdmin && (
                <div className="modal-backdrop">
                    <div className="modal modal-wide">
                        <div className="modal-header">
                            <h3>
                                {editingBookId ? "Изменить книгу" : "Новая книга"}
                            </h3>
                            <button
                                className="icon-button close-button"
                                type="button"
                                onClick={() => {
                                    setIsBookModalOpen(false)
                                    setEditingBookId(null)
                                }}
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
                                                    title={publisher.name}
                                                >
                                                    {truncateLabel(publisher.name)}
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
                                <div className="field-label">
                                    <span>Обложка</span>
                                    <div className="file-input">
                                        <input
                                            id="book-cover"
                                            type="file"
                                            accept="image/*"
                                            onChange={(event) => {
                                                const file =
                                                    event.target.files?.[0] ?? null
                                                if (coverPreview) {
                                                    URL.revokeObjectURL(
                                                        coverPreview
                                                    )
                                                }
                                                if (!file) {
                                                    setCoverFile(null)
                                                    setCoverPreview(null)
                                                    setCoverFileName("")
                                                    return
                                                }
                                                const url =
                                                    URL.createObjectURL(file)
                                                setCoverFile(file)
                                                setCoverPreview(url)
                                                setCoverFileName(file.name)
                                            }}
                                        />
                                        <label
                                            className="ghost-button file-button"
                                            htmlFor="book-cover"
                                        >
                                            Выбрать
                                        </label>
                                        <span className="item-meta file-name">
                                            {coverFileName || "Файл не выбран"}
                                        </span>
                                    </div>
                                    <span className="item-meta">
                                        Файл будет загружен после сохранения книги.
                                    </span>
                                    {coverPreview && (
                                        <img
                                            className="cover-preview"
                                            src={coverPreview}
                                            alt="Предпросмотр обложки"
                                        />
                                    )}
                                    {!coverPreview &&
                                        editingBookId &&
                                        selectedBook && (
                                            <SafeImage
                                                src={getCoverUrl(selectedBook)}
                                                alt={selectedBook.title}
                                                className="cover-preview"
                                            />
                                        )}
                                </div>
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
                                            onClick={openWorkCreate}
                                            aria-label="Добавить произведение"
                                        >
                                            +
                                        </button>
                                    </div>
                                </div>
                                <div className="work-picker">
                                    <label className="field-label">
                                        Поиск
                                        <input
                                            className="text-input"
                                            placeholder="Начните вводить название"
                                            value={workSearch}
                                            onChange={(event) =>
                                                setWorkSearch(event.target.value)
                                            }
                                        />
                                    </label>
                                    <div className="work-picker-columns">
                                        <div className="work-picker-panel">
                                            <h5>Найденные произведения</h5>
                                            <div className="stack">
                                                {works
                                                    .filter((work) => {
                                                        if (
                                                            bookDraft.workIds.includes(
                                                                work.id
                                                            )
                                                        ) {
                                                            return false
                                                        }
                                                        const query =
                                                            workSearch.trim().toLowerCase()
                                                        if (!query) {
                                                            return true
                                                        }
                                                        return work.title
                                                            .toLowerCase()
                                                            .includes(query)
                                                    })
                                                    .map((work) => (
                                                        <div
                                                            key={work.id}
                                                            className="work-picker-row"
                                                        >
                                                            <div>
                                                                <div className="work-title">
                                                                    {work.title}
                                                                </div>
                                                                <small className="item-meta">
                                                                    {(() => {
                                                                        const names = (
                                                                            work.authors ??
                                                                            []
                                                                        )
                                                                            .map(getAuthorName)
                                                                            .join(", ")
                                                                        const base =
                                                                            names ||
                                                                            "Без автора"
                                                                        return work.year
                                                                            ? `${base}, ${work.year}`
                                                                            : base
                                                                    })()}
                                                                </small>
                                                            </div>
                                                            <button
                                                                className="ghost-button"
                                                                type="button"
                                                                onClick={() =>
                                                                    setBookDraft(
                                                                        (prev) => ({
                                                                            ...prev,
                                                                            workIds: [
                                                                                ...prev.workIds,
                                                                                work.id,
                                                                            ],
                                                                        })
                                                                    )
                                                                }
                                                            >
                                                                Добавить
                                                            </button>
                                                        </div>
                                                    ))}
                                                {works.filter((work) => {
                                                    if (
                                                        bookDraft.workIds.includes(
                                                            work.id
                                                        )
                                                    ) {
                                                        return false
                                                    }
                                                    const query =
                                                        workSearch.trim().toLowerCase()
                                                    if (!query) {
                                                        return true
                                                    }
                                                    return work.title
                                                        .toLowerCase()
                                                        .includes(query)
                                                }).length === 0 && (
                                                    <p className="status-line">
                                                        Ничего не найдено
                                                    </p>
                                                )}
                                            </div>
                                        </div>
                                        <div className="work-picker-panel">
                                            <h5>Выбранные произведения</h5>
                                            <div className="stack">
                                                {bookDraft.workIds.length === 0 && (
                                                    <p className="status-line">
                                                        Пока нет выбранных
                                                    </p>
                                                )}
                                                {bookDraft.workIds.map((workId) => {
                                                    const work = works.find(
                                                        (item) =>
                                                            item.id === workId
                                                    )
                                                    if (!work) {
                                                        return null
                                                    }
                                                    return (
                                                        <div
                                                            key={work.id}
                                                            className="work-picker-row"
                                                        >
                                                            <div>
                                                                <div className="work-title">
                                                                    {work.title}
                                                                </div>
                                                                <small className="item-meta">
                                                                    {(() => {
                                                                        const names = (
                                                                            work.authors ??
                                                                            []
                                                                        )
                                                                            .map(getAuthorName)
                                                                            .join(", ")
                                                                        const base =
                                                                            names ||
                                                                            "Без автора"
                                                                        return work.year
                                                                            ? `${base}, ${work.year}`
                                                                            : base
                                                                    })()}
                                                                </small>
                                                            </div>
                                                            <button
                                                                className="ghost-button"
                                                                type="button"
                                                                onClick={() =>
                                                                    setBookDraft(
                                                                        (prev) => ({
                                                                            ...prev,
                                                                            workIds:
                                                                                prev.workIds.filter(
                                                                                    (id) =>
                                                                                        id !==
                                                                                        work.id
                                                                                ),
                                                                        })
                                                                    )
                                                                }
                                                            >
                                                                Убрать
                                                            </button>
                                                        </div>
                                                    )
                                                })}
                                            </div>
                                        </div>
                                    </div>
                                </div>
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
                                {bookSaving
                                    ? "Сохранение..."
                                    : editingBookId
                                    ? "Сохранить"
                                    : "Добавить книгу"}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {isBookInfoOpen && selectedBook && (
                <div className="modal-backdrop">
                    <div className="modal modal-wide book-info-modal">
                        <div className="modal-header">
                            <div className="book-info-header">
                                <SafeImage
                                    src={getCoverUrl(selectedBook)}
                                    alt={selectedBook.title}
                                    className="cover-preview book-info-cover"
                                />
                                <div className="book-info-header-text">
                                    <h3>{selectedBook.title}</h3>
                                    <div className="book-info-header-meta">
                                        <p>
                                            {selectedBook.works?.length
                                                ? Array.from(
                                                      new Set(
                                                          selectedBook.works.flatMap(
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
                                        {selectedBook.publisher?.name && (
                                            <p className="item-meta">
                                                {selectedBook.publisher.name}
                                            </p>
                                        )}
                                        {selectedBook.year && (
                                            <p className="item-meta">
                                                {selectedBook.year}
                                            </p>
                                        )}
                                    </div>
                                </div>
                            </div>
                            <button
                                className="icon-button close-button"
                                type="button"
                                onClick={() => setIsBookInfoOpen(false)}
                            >
                                ✕
                            </button>
                        </div>
                        <div className="modal-body">
                            <div className="stack book-info-stack">
                                <div className="book-info-meta">
                                    <p className="item-meta">
                                        Локация:{" "}
                                        {"location" in selectedBook
                                            ? formatLocation(
                                                  (selectedBook as BookInternal)
                                                      .location
                                              )
                                            : "—"}
                                    </p>
                                </div>
                                {selectedBook.description && (
                                    <p>{selectedBook.description}</p>
                                )}
                                <div className="stack">
                                    <h4 className="subheading">Произведения</h4>
                                    {(selectedBook.works ?? []).length === 0 && (
                                        <p className="status-line">Нет данных</p>
                                    )}
                                    <div className="stack book-info-works">
                                        {(selectedBook.works ?? []).map((work) => (
                                            <button
                                                key={work.id}
                                                className="mini-card list-item-selectable"
                                                type="button"
                                                onClick={async () => {
                                                    setActiveTab("works")
                                                    await handleWorkSelect(
                                                        work,
                                                        false
                                                    )
                                                    setIsBookInfoOpen(false)
                                                }}
                                            >
                                                <span>{work.title}</span>
                                                <span className="item-meta">
                                                    {(() => {
                                                        const names = (
                                                            work.authors ?? []
                                                        )
                                                            .map(getAuthorName)
                                                            .join(", ")
                                                        return names || "Без автора"
                                                    })()}
                                                </span>
                                            </button>
                                        ))}
                                    </div>
                                </div>
                                <div className="book-info-timestamps">
                                    {(selectedBook as BookPublic).created_at && (
                                        <p className="item-meta">
                                            Создано:{" "}
                                            {formatDateTime(
                                                (selectedBook as BookPublic).created_at
                                            )}
                                        </p>
                                    )}
                                    {(selectedBook as BookPublic).updated_at && (
                                        <p className="item-meta">
                                            Обновлено:{" "}
                                            {formatDateTime(
                                                (selectedBook as BookPublic).updated_at
                                            )}
                                        </p>
                                    )}
                                </div>
                            </div>
                        </div>
                        {isAdmin && (
                            <div className="modal-footer">
                                <button
                                    className="ghost-button"
                                    type="button"
                                    onClick={() =>
                                        addBookToPrintQueue(selectedBook)
                                    }
                                >
                                    В очередь печати
                                </button>
                                <button
                                    className="ghost-button"
                                    type="button"
                                    onClick={() => openBookEdit(selectedBook)}
                                >
                                    Изменить
                                </button>
                            </div>
                        )}
                    </div>
                </div>
            )}

            {isWorkModalOpen && (
                <div className="modal-backdrop">
                    <div className="modal">
                        <div className="modal-header">
                            <h3>
                                {editingWorkId
                                    ? "Изменить произведение"
                                    : "Новое произведение"}
                            </h3>
                            <button
                                className="icon-button close-button"
                                type="button"
                                onClick={closeWorkModal}
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
                                <div className="work-picker">
                                    <label className="field-label">
                                        Поиск
                                        <input
                                            className="text-input"
                                            placeholder="Начните вводить имя"
                                            value={workAuthorSearch}
                                            onChange={(event) =>
                                                setWorkAuthorSearch(
                                                    event.target.value
                                                )
                                            }
                                        />
                                    </label>
                                    <div className="work-picker-columns">
                                        <div className="work-picker-panel">
                                            <h5>Найденные авторы</h5>
                                            <div className="stack">
                                                {authors
                                                    .filter((author) => {
                                                        if (
                                                            workDraft.authorIds.includes(
                                                                author.id
                                                            )
                                                        ) {
                                                            return false
                                                        }
                                                        const query =
                                                            workAuthorSearch
                                                                .trim()
                                                                .toLowerCase()
                                                        if (!query) {
                                                            return true
                                                        }
                                                        return getAuthorName(
                                                            author
                                                        )
                                                            .toLowerCase()
                                                            .includes(query)
                                                    })
                                                    .map((author) => (
                                                        <div
                                                            key={author.id}
                                                            className="work-picker-row"
                                                        >
                                                            <div>
                                                                <div className="work-title">
                                                                    {getAuthorName(
                                                                        author
                                                                    )}
                                                                </div>
                                                            </div>
                                                            <button
                                                                className="ghost-button"
                                                                type="button"
                                                                onClick={() =>
                                                                    setWorkDraft(
                                                                        (prev) => ({
                                                                            ...prev,
                                                                            authorIds:
                                                                                [
                                                                                    ...prev.authorIds,
                                                                                    author.id,
                                                                                ],
                                                                        })
                                                                    )
                                                                }
                                                            >
                                                                Добавить
                                                            </button>
                                                        </div>
                                                    ))}
                                                {authors.filter((author) => {
                                                    if (
                                                        workDraft.authorIds.includes(
                                                            author.id
                                                        )
                                                    ) {
                                                        return false
                                                    }
                                                    const query =
                                                        workAuthorSearch
                                                            .trim()
                                                            .toLowerCase()
                                                    if (!query) {
                                                        return true
                                                    }
                                                    return getAuthorName(author)
                                                        .toLowerCase()
                                                        .includes(query)
                                                }).length === 0 && (
                                                    <p className="status-line">
                                                        Ничего не найдено
                                                    </p>
                                                )}
                                            </div>
                                        </div>
                                        <div className="work-picker-panel">
                                            <h5>Выбранные авторы</h5>
                                            <div className="stack">
                                                {workDraft.authorIds.length ===
                                                    0 && (
                                                    <p className="status-line">
                                                        Пока нет выбранных
                                                    </p>
                                                )}
                                                {workDraft.authorIds.map(
                                                    (authorId) => {
                                                        const author =
                                                            authors.find(
                                                                (item) =>
                                                                    item.id ===
                                                                    authorId
                                                            )
                                                        if (!author) {
                                                            return null
                                                        }
                                                        return (
                                                            <div
                                                                key={author.id}
                                                                className="work-picker-row"
                                                            >
                                                                <div>
                                                                    <div className="work-title">
                                                                        {getAuthorName(
                                                                            author
                                                                        )}
                                                                    </div>
                                                                </div>
                                                                <button
                                                                    className="ghost-button"
                                                                    type="button"
                                                                    onClick={() =>
                                                                        setWorkDraft(
                                                                            (prev) => ({
                                                                                ...prev,
                                                                                authorIds:
                                                                                    prev.authorIds.filter(
                                                                                        (
                                                                                            id
                                                                                        ) =>
                                                                                            id !==
                                                                                            author.id
                                                                                    ),
                                                                            })
                                                                        )
                                                                    }
                                                                >
                                                                    Убрать
                                                                </button>
                                                            </div>
                                                        )
                                                    }
                                                )}
                                            </div>
                                        </div>
                                    </div>
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
                                    : editingWorkId
                                    ? "Сохранить"
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
                                className="icon-button close-button"
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
                                <div className="field-label">
                                    <span>Фото</span>
                                    <div className="file-input">
                                        <input
                                            id="author-photo"
                                            type="file"
                                            accept="image/*"
                                            onChange={(event) => {
                                                const file =
                                                    event.target.files?.[0] ?? null
                                                if (authorPhotoPreview) {
                                                    URL.revokeObjectURL(
                                                        authorPhotoPreview
                                                    )
                                                }
                                                if (!file) {
                                                    setAuthorPhotoFile(null)
                                                    setAuthorPhotoPreview(null)
                                                    return
                                                }
                                                const url =
                                                    URL.createObjectURL(file)
                                                setAuthorPhotoFile(file)
                                                setAuthorPhotoPreview(url)
                                                setAuthorDraft((prev) => ({
                                                    ...prev,
                                                    photoUrl: "",
                                                }))
                                            }}
                                        />
                                        <label
                                            className="ghost-button file-button"
                                            htmlFor="author-photo"
                                        >
                                            Выбрать
                                        </label>
                                        <span className="item-meta file-name">
                                            {authorPhotoFile?.name ||
                                                "Файл не выбран"}
                                        </span>
                                    </div>
                                    {authorPhotoPreview && (
                                        <img
                                            className="avatar"
                                            src={authorPhotoPreview}
                                            alt="Предпросмотр фото автора"
                                        />
                                    )}
                                </div>
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
                                className="icon-button close-button"
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
                            <div className="field-label">
                                <span>Логотип</span>
                                <div className="file-input">
                                    <input
                                        id="publisher-logo"
                                        type="file"
                                        accept="image/*"
                                        onChange={(event) => {
                                            const file =
                                                event.target.files?.[0] ?? null
                                            if (publisherLogoPreview) {
                                                URL.revokeObjectURL(
                                                    publisherLogoPreview
                                                )
                                            }
                                            if (!file) {
                                                setPublisherLogoFile(null)
                                                setPublisherLogoPreview(null)
                                                return
                                            }
                                            const url = URL.createObjectURL(file)
                                            setPublisherLogoFile(file)
                                            setPublisherLogoPreview(url)
                                            setPublisherDraft((prev) => ({
                                                ...prev,
                                                logoUrl: "",
                                            }))
                                        }}
                                    />
                                    <label
                                        className="ghost-button file-button"
                                        htmlFor="publisher-logo"
                                    >
                                        Выбрать
                                    </label>
                                    <span className="item-meta file-name">
                                        {publisherLogoFile?.name ||
                                            "Файл не выбран"}
                                    </span>
                                </div>
                                {publisherLogoPreview && (
                                    <img
                                        className="avatar"
                                        src={publisherLogoPreview}
                                        alt="Предпросмотр логотипа"
                                    />
                                )}
                            </div>
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

            {isPublisherInfoOpen && (
                <div className="modal-backdrop">
                    <div className="modal">
                        <div className="modal-header">
                            <div>
                                <h3>
                                    {selectedPublisher
                                        ? selectedPublisher.name
                                        : "Загрузка..."}
                                </h3>
                            </div>
                            <button
                                className="icon-button close-button"
                                type="button"
                                onClick={() => setIsPublisherInfoOpen(false)}
                            >
                                ✕
                            </button>
                        </div>
                        <div className="modal-body">
                            {selectedPublisher ? (
                                <div className="stack">
                                    {selectedPublisher.logo_url && (
                                        <SafeImage
                                            src={selectedPublisher.logo_url}
                                            alt={selectedPublisher.name}
                                            className="publisher-logo-full"
                                        />
                                    )}
                                    <div className="stack author-info-stack">
                                        {selectedPublisher.web_url && (
                                            <p>
                                                <strong>Сайт:</strong>{" "}
                                                {selectedPublisher.web_url}
                                            </p>
                                        )}
                                    </div>
                                </div>
                            ) : (
                                <p className="status-line">Загрузка...</p>
                            )}
                        </div>
                        {isAdmin && (
                            <div className="modal-footer modal-footer-actions">
                                {publisherInfoError && (
                                    <p className="error-banner">
                                        {publisherInfoError}
                                    </p>
                                )}
                                <button
                                    className="ghost-button"
                                    type="button"
                                    onClick={() =>
                                        selectedPublisher &&
                                        openPublisherEdit(selectedPublisher)
                                    }
                                    disabled={!selectedPublisher || publisherInfoSaving}
                                >
                                    Изменить
                                </button>
                                <button
                                    className="primary-button"
                                    type="button"
                                    onClick={handlePublisherDelete}
                                    disabled={!selectedPublisher || publisherInfoSaving}
                                >
                                    Удалить
                                </button>
                            </div>
                        )}
                    </div>
                </div>
            )}

            {isPublisherEditOpen && isAdmin && (
                <div className="modal-backdrop">
                    <div className="modal">
                        <div className="modal-header">
                            <h3>Изменить издательство</h3>
                            <button
                                className="icon-button close-button"
                                type="button"
                                onClick={() => setIsPublisherEditOpen(false)}
                            >
                                ✕
                            </button>
                        </div>
                        <div className="modal-body">
                            {selectedPublisher?.logo_url && (
                                <SafeImage
                                    src={selectedPublisher.logo_url}
                                    alt={selectedPublisher.name}
                                    className="publisher-logo-full"
                                />
                            )}
                            <label className="field-label">
                                Название
                                <input
                                    className="text-input"
                                    value={publisherEditDraft.name}
                                    onChange={(event) =>
                                        setPublisherEditDraft((prev) => ({
                                            ...prev,
                                            name: event.target.value,
                                        }))
                                    }
                                />
                            </label>
                            <label className="field-label">
                                Сайт
                                <input
                                    className="text-input"
                                    value={publisherEditDraft.webUrl}
                                    onChange={(event) =>
                                        setPublisherEditDraft((prev) => ({
                                            ...prev,
                                            webUrl: event.target.value,
                                        }))
                                    }
                                />
                            </label>
                            <div className="field-label">
                                <span>Логотип</span>
                                <div className="file-input">
                                    <input
                                        id="publisher-logo-edit"
                                        type="file"
                                        accept="image/*"
                                        onChange={(event) => {
                                            const file =
                                                event.target.files?.[0] ?? null
                                            if (publisherEditLogoPreview) {
                                                URL.revokeObjectURL(
                                                    publisherEditLogoPreview
                                                )
                                            }
                                            if (!file) {
                                                setPublisherEditLogoFile(null)
                                                setPublisherEditLogoPreview(null)
                                                return
                                            }
                                            const url = URL.createObjectURL(file)
                                            setPublisherEditLogoFile(file)
                                            setPublisherEditLogoPreview(url)
                                        }}
                                    />
                                    <label
                                        className="ghost-button file-button"
                                        htmlFor="publisher-logo-edit"
                                    >
                                        Выбрать
                                    </label>
                                    <span className="item-meta file-name">
                                        {publisherEditLogoFile?.name ||
                                            "Файл не выбран"}
                                    </span>
                                </div>
                                {publisherEditLogoPreview && (
                                    <img
                                        className="avatar"
                                        src={publisherEditLogoPreview}
                                        alt="Предпросмотр логотипа"
                                    />
                                )}
                            </div>
                            {publisherEditError && (
                                <p className="error-banner">{publisherEditError}</p>
                            )}
                        </div>
                        <div className="modal-footer">
                            <button
                                className="primary-button"
                                type="button"
                                onClick={handlePublisherEditSave}
                                disabled={publisherEditSaving}
                            >
                                {publisherEditSaving
                                    ? "Сохранение..."
                                    : "Сохранить"}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {isAuthorEditOpen && isAdmin && (
                <div className="modal-backdrop modal-backdrop-top">
                    <div className="modal">
                        <div className="modal-header">
                            <h3>Изменить автора</h3>
                            <button
                                className="icon-button close-button"
                                type="button"
                                onClick={() => setIsAuthorEditOpen(false)}
                            >
                                ✕
                            </button>
                        </div>
                        <div className="modal-body">
                            {selectedAuthor?.photo_url && (
                                <SafeImage
                                    src={selectedAuthor.photo_url}
                                    alt={getAuthorName(selectedAuthor)}
                                    className="avatar"
                                />
                            )}
                            <label className="field-label">
                                Фамилия
                                <input
                                    className="text-input"
                                    value={authorEditDraft.lastName}
                                    onChange={(event) =>
                                        setAuthorEditDraft((prev) => ({
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
                                    value={authorEditDraft.firstName}
                                    onChange={(event) =>
                                        setAuthorEditDraft((prev) => ({
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
                                    value={authorEditDraft.middleName}
                                    onChange={(event) =>
                                        setAuthorEditDraft((prev) => ({
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
                                    value={authorEditDraft.birthDate}
                                    onChange={(event) =>
                                        setAuthorEditDraft((prev) => ({
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
                                    value={authorEditDraft.deathDate}
                                    onChange={(event) =>
                                        setAuthorEditDraft((prev) => ({
                                            ...prev,
                                            deathDate: event.target.value,
                                        }))
                                    }
                                />
                            </label>
                            <div className="field-label">
                                <span>Фото</span>
                                <div className="file-input">
                                    <input
                                        id="author-photo-edit"
                                        type="file"
                                        accept="image/*"
                                        onChange={(event) => {
                                            const file =
                                                event.target.files?.[0] ?? null
                                            if (authorEditPhotoPreview) {
                                                URL.revokeObjectURL(
                                                    authorEditPhotoPreview
                                                )
                                            }
                                            if (!file) {
                                                setAuthorEditPhotoFile(null)
                                                setAuthorEditPhotoPreview(null)
                                                return
                                            }
                                            const url = URL.createObjectURL(file)
                                            setAuthorEditPhotoFile(file)
                                            setAuthorEditPhotoPreview(url)
                                        }}
                                    />
                                    <label
                                        className="ghost-button file-button"
                                        htmlFor="author-photo-edit"
                                    >
                                        Выбрать
                                    </label>
                                    <span className="item-meta file-name">
                                        {authorEditPhotoFile?.name ||
                                            "Файл не выбран"}
                                    </span>
                                </div>
                                {authorEditPhotoPreview && (
                                    <img
                                        className="avatar"
                                        src={authorEditPhotoPreview}
                                        alt="Предпросмотр фото автора"
                                    />
                                )}
                            </div>
                            <label className="field-label full">
                                Биография
                                <textarea
                                    className="text-area"
                                    value={authorEditDraft.bio}
                                    onChange={(event) =>
                                        setAuthorEditDraft((prev) => ({
                                            ...prev,
                                            bio: event.target.value,
                                        }))
                                    }
                                />
                            </label>
                            {authorEditError && (
                                <p className="error-banner">{authorEditError}</p>
                            )}
                        </div>
                        <div className="modal-footer">
                            <button
                                className="primary-button"
                                type="button"
                                onClick={handleAuthorEditSave}
                                disabled={authorEditSaving}
                            >
                                {authorEditSaving
                                    ? "Сохранение..."
                                    : "Сохранить"}
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
                                className="icon-button close-button"
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
                                            address:
                                                nextType === "building"
                                                    ? prev.address
                                                    : "",
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
                            {getParentType(locationDraft.type) && (
                                <label className="field-label">
                                    Родитель
                                    <select
                                        className={`text-input ${
                                            locationDraft.parentId
                                                ? ""
                                                : "select-empty"
                                        }`}
                                        value={locationDraft.parentId}
                                        onChange={(event) =>
                                            setLocationDraft((prev) => ({
                                                ...prev,
                                                parentId: event.target.value,
                                            }))
                                        }
                                        disabled={locationDraft.lockParent}
                                    >
                                        <option value="">
                                            Не выбрано
                                        </option>
                                        {(
                                            locationByType[
                                                getParentType(
                                                    locationDraft.type
                                                )!
                                            ] ?? []
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
                            )}
                            {locationDraft.type === "building" && (
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
                            )}
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
                            <div>
                                <h3>
                                    {selectedAuthor
                                        ? getAuthorName(selectedAuthor)
                                        : "Загрузка..."}
                                </h3>
                                {selectedAuthor && (
                                    <p className="item-meta author-life">
                                        {formatLifeDates(selectedAuthor)}
                                    </p>
                                )}
                            </div>
                            <button
                                className="icon-button close-button"
                                type="button"
                                onClick={() => setIsAuthorInfoOpen(false)}
                            >
                                ✕
                            </button>
                        </div>
                        <div className="modal-body">
                            {selectedAuthor ? (
                                <div className="stack">
                                    {selectedAuthor.photo_url && (
                                        <SafeImage
                                            src={selectedAuthor.photo_url}
                                            alt={getAuthorName(selectedAuthor)}
                                            className="publisher-logo-full"
                                        />
                                    )}
                                    <div>
                                        <div className="stack author-info-stack">
                                            {selectedAuthor.bio && (
                                                <div className="author-bio-scroll">
                                                    <p>
                                                        <strong>Биография:</strong>{" "}
                                                        {selectedAuthor.bio}
                                                    </p>
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                </div>
                            ) : (
                                <p className="status-line">Загрузка...</p>
                            )}
                        </div>
                        {isAdmin && (
                            <div className="modal-footer modal-footer-actions">
                                {authorInfoError && (
                                    <p className="error-banner">
                                        {authorInfoError}
                                    </p>
                                )}
                                <button
                                    className="ghost-button"
                                    type="button"
                                    onClick={() =>
                                        selectedAuthor &&
                                        openAuthorEdit(selectedAuthor)
                                    }
                                    disabled={!selectedAuthor || authorInfoSaving}
                                >
                                    Изменить
                                </button>
                            </div>
                        )}
                    </div>
                </div>
            )}

            {isWorkInfoOpen && (
                <div className="modal-backdrop">
                    <div className="modal">
                        <div className="modal-header">
                            <h3>Произведение</h3>
                            <button
                                className="icon-button close-button"
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
                                    onClick={handleDeleteWork}
                                    disabled={!selectedWorkDetail}
                                >
                                    Удалить
                                </button>
                                <button
                                    className="ghost-button"
                                    type="button"
                                    onClick={openWorkEditFromSelection}
                                    disabled={!selectedWorkDetail}
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

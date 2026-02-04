import {useEffect, useState} from "react"
import {API_URL} from "../api/http"

type Props = {
    src?: string | null
    alt: string
    className?: string
}

export function SafeImage({src, alt, className}: Props) {
    const [hidden, setHidden] = useState(false)
    const [isOpen, setIsOpen] = useState(false)
    const [zoom, setZoom] = useState(1)
    const [offset, setOffset] = useState({x: 0, y: 0})
    const [isDragging, setIsDragging] = useState(false)
    const [dragStart, setDragStart] = useState({x: 0, y: 0})

    useEffect(() => {
        if (!isOpen) {
            return
        }
        const onKeyDown = (event: KeyboardEvent) => {
            if (event.key === "Escape") {
                setIsOpen(false)
            }
        }
        document.addEventListener("keydown", onKeyDown)
        const originalOverflow = document.body.style.overflow
        document.body.style.overflow = "hidden"
        return () => {
            document.removeEventListener("keydown", onKeyDown)
            document.body.style.overflow = originalOverflow
        }
    }, [isOpen])

    if (!src || hidden) {
        return (
            <div
                className={["image-placeholder", className]
                    .filter(Boolean)
                    .join(" ")}
                aria-label={alt}
            />
        )
    }

    const resolvedSrc = resolveImageSrc(src)

    const handleOpen = () => {
        setZoom(1)
        setOffset({x: 0, y: 0})
        setIsOpen(true)
    }

    const handleWheel = (event: React.WheelEvent) => {
        event.preventDefault()
        const delta = event.deltaY > 0 ? -0.1 : 0.1
        setZoom((prev) => Math.min(4, Math.max(1, prev + delta)))
    }

    const handlePointerDown = (event: React.PointerEvent) => {
        if (zoom <= 1) {
            return
        }
        event.preventDefault()
        setIsDragging(true)
        event.currentTarget.setPointerCapture(event.pointerId)
        setDragStart({
            x: event.clientX - offset.x,
            y: event.clientY - offset.y,
        })
    }

    const handlePointerMove = (event: React.PointerEvent) => {
        if (!isDragging) {
            return
        }
        event.preventDefault()
        setOffset({
            x: event.clientX - dragStart.x,
            y: event.clientY - dragStart.y,
        })
    }

    const handlePointerUp = (event: React.PointerEvent) => {
        setIsDragging(false)
        try {
            event.currentTarget.releasePointerCapture(event.pointerId)
        } catch {
            // ignore release errors
        }
    }

    return (
        <>
            <img
                src={resolvedSrc}
                alt={alt}
                className={className}
                onError={() => setHidden(true)}
                onClick={handleOpen}
                role="button"
                tabIndex={0}
                onKeyDown={(event) => {
                    if (event.key === "Enter" || event.key === " ") {
                        event.preventDefault()
                        handleOpen()
                    }
                }}
            />
            {isOpen && (
                <div
                    className="image-viewer-backdrop"
                    onClick={() => setIsOpen(false)}
                >
                    <div
                        className="image-viewer"
                        onClick={(event) => event.stopPropagation()}
                    >
                        <div className="image-viewer-toolbar">
                            <button
                                className="ghost-button"
                                type="button"
                                onClick={() =>
                                    setZoom((prev) => Math.max(1, prev - 0.2))
                                }
                            >
                                −
                            </button>
                            <button
                                className="ghost-button"
                                type="button"
                                onClick={() => setZoom(1)}
                            >
                                100%
                            </button>
                            <button
                                className="ghost-button"
                                type="button"
                                onClick={() =>
                                    setZoom((prev) => Math.min(4, prev + 0.2))
                                }
                            >
                                +
                            </button>
                            <button
                                className="ghost-button"
                                type="button"
                                onClick={() => setOffset({x: 0, y: 0})}
                            >
                                Центр
                            </button>
                            <button
                                className="ghost-button"
                                type="button"
                                onClick={() => setIsOpen(false)}
                            >
                                Закрыть
                            </button>
                        </div>
                        <div
                            className="image-viewer-canvas"
                            onWheel={handleWheel}
                            onPointerDown={handlePointerDown}
                            onPointerMove={handlePointerMove}
                            onPointerUp={handlePointerUp}
                            onPointerLeave={handlePointerUp}
                        >
                            <img
                                src={resolvedSrc}
                                alt={alt}
                                className="image-viewer-img"
                                style={{
                                    transform: `translate(${offset.x}px, ${offset.y}px) scale(${zoom})`,
                                    cursor:
                                        zoom > 1
                                            ? isDragging
                                                ? "grabbing"
                                                : "grab"
                                            : "default",
                                }}
                                draggable={false}
                                onDragStart={(event) => event.preventDefault()}
                            />
                        </div>
                    </div>
                </div>
            )}
        </>
    )
}

function resolveImageSrc(src: string) {
    if (!src.startsWith("/")) {
        return src
    }
    if (!src.startsWith("/static/")) {
        return src
    }
    const apiBase = API_URL.replace(/\/$/, "")
    if (!apiBase) {
        return src
    }
    return `${apiBase}${src}`
}

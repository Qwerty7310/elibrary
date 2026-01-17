import {useState} from "react"

type Props = {
    src?: string | null
    alt: string
    className?: string
}

export function SafeImage({src, alt, className}: Props) {
    const [hidden, setHidden] = useState(false)

    if (!src || hidden) {
        return null
    }

    return (
        <img
            src={src}
            alt={alt}
            className={className}
            onError={() => setHidden(true)}
        />
    )
}


"use client"

// Rasterizes an inline SVG chart to a downloadable PNG — the "take a
// screenshot of graphs" feature from the spec. No new dependency: draw
// the serialized SVG into an <img>, then into a canvas, then export.
export function downloadSvgAsPng(svg: SVGSVGElement, filename: string) {
    const width = Number(svg.getAttribute("width")) || svg.clientWidth
    const height = Number(svg.getAttribute("height")) || svg.clientHeight

    const svgString = new XMLSerializer().serializeToString(svg)
    const svgBlob = new Blob([svgString], { type: "image/svg+xml;charset=utf-8" })
    const url = URL.createObjectURL(svgBlob)

    const image = new Image()
    image.onload = () => {
        const scale = 2 // export at 2x for a crisp screenshot
        const canvas = document.createElement("canvas")
        canvas.width = width * scale
        canvas.height = height * scale

        const ctx = canvas.getContext("2d")
        URL.revokeObjectURL(url)
        if (!ctx) return

        ctx.scale(scale, scale)
        // getComputedStyle so the exported PNG matches whichever theme
        // (light/dark) the chart is currently rendered in.
        ctx.fillStyle = getComputedStyle(document.body).backgroundColor || "#ffffff"
        ctx.fillRect(0, 0, width, height)
        ctx.drawImage(image, 0, 0, width, height)

        canvas.toBlob((blob) => {
            if (!blob) return
            const link = document.createElement("a")
            link.href = URL.createObjectURL(blob)
            link.download = filename
            link.click()
            URL.revokeObjectURL(link.href)
        })
    }
    image.src = url
}

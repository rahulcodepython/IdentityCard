import { NextResponse } from "next/server"

export const NotFoundResponse = (message: string) => {
    return NextResponse.json(
        { success: false, error: message },
        { status: 404 }
    )
}

export const BadRequestResponse = (message: string) => {
    return NextResponse.json(
        { success: false, error: message },
        { status: 400 }
    )
}

export const InternalServerErrorResponse = (message: string) => {
    return NextResponse.json(
        { success: false, error: message },
        { status: 500 }
    )
}

export const SuccessResponse = <T>(data: T) => {
    return NextResponse.json(
        { success: true, data },
        { status: 200 }
    )
}

export const UnauthorizedResponse = (message: string) => {
    return NextResponse.json(
        { success: false, error: message },
        { status: 401 }
    )
}

export const CreatedResponse = <T>(data: T) => {
    return NextResponse.json(
        { success: true, data },
        { status: 201 }
    )
}
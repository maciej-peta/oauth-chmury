import { NextResponse } from 'next/server'

export async function GET() {
    console.log('Healthcheck: frontend - ok')
    return NextResponse.json({ msg: 'ok' })
}

"use client"

import {logout} from "@/lib/auth"

export default function LoginComponent(){

    return <button onClick={() => logout()}> sign out</button>
}

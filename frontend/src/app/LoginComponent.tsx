"use client"

import {login} from "@/lib/auth"

export default function LoginComponent(){

    return <button onClick={() => login()}> sign up</button>
}

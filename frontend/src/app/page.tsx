"use server"

import {auth} from "@/auth"
import LoginComponent from "./LoginComponent";
import MainComponent from "@/app/MainComponent";
import Link from "next/link";

export default async function Home() {
    const session = await auth();
    console.log(session?.accessToken)

    if(!session?.user){
        return <LoginComponent></LoginComponent>
    }
    return (<div>
        <Link href="/user-info">link</Link>
        <MainComponent user={{
            name: session.user.name,
            email: session.user.email,
            picture: session.user.image,
            sub: session.user.id
        }} accessToken={session.accessToken}></MainComponent>
    </div>);
}
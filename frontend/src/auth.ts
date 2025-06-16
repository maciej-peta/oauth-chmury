// app/auth.ts
import NextAuth from "next-auth";
import Auth0Provider from "next-auth/providers/auth0";

// Core NextAuth setup
export const { auth, handlers, signIn, signOut } = NextAuth({
    providers: [
        Auth0Provider({
            clientId: process.env.AUTH0_CLIENT_ID!,
            clientSecret: process.env.AUTH0_CLIENT_SECRET!,
            issuer: `https://${process.env.AUTH0_DOMAIN}`,
            authorization: {
                params: {
                    audience: "https://file-conversion-api/", //todo: move api path
                    scope: "openid profile email",
                },
            },
        }),
    ],
    session: {
        strategy: "jwt",
    },
    callbacks: {
        async jwt({ token, account, profile }) {
            if (account) {
                token.accessToken = account.access_token;

                token.id = profile?.sub ?? undefined;
            }
            return token;
        },
        async session({ session, token }) {
            session.accessToken = token.accessToken;
            session.user.id = token.id ?? ""
            return session;
        },
    },
});

"use client";

export default function RootLayout({ children }: { children: React.ReactNode }) {
    return (
        <html lang="en">
        <head>{/* ... any metadata ... */}</head>
        <body>
        {children}
        </body>
        </html>
    );
}

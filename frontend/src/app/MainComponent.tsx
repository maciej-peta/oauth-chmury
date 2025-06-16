"use client";

import { useState, useEffect, useCallback, DragEvent, ChangeEvent } from "react";
import {logout} from "@/lib/auth"

type MainComponentProps = {
    user: {
        name?: string | null;
        email?: string | null;
        picture?: string | null;
        sub?: string | null;
        [key: string]: string | null | undefined;
    };
    accessToken: string | undefined;
};

export default function MainComponent({ user, accessToken }: MainComponentProps) {
    const [file, setFile] = useState<File | null>(null);
    const [isDragging, setIsDragging] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [isConverting, setIsConverting] = useState(false);
    const [startingType, setStartingType] = useState<string>("image/jpeg");
    const [targetType, setTargetType] = useState<string>("image/jpeg");

    const allowedTypes: string[] = ["image/jpeg", "image/png", "image/webp"];
    const backendAddress: string = process.env.NEXT_PUBLIC_BACKEND_URL || "";

    useEffect(() => {
        const checkAndRegisterUser = async () => {
            if (!user?.sub) return;

            try {
                const userUrl = `${backendAddress}/users/${user.sub}`;
                const getResponse = await fetch(userUrl, {
                    headers: {
                        "Authorization": `Bearer ${accessToken}`,
                    },
                });

                if (getResponse.status === 404) {
                    console.log("User not found. Registering...");
                    const postResponse = await fetch(`${backendAddress}/users`, {
                        method: "POST",
                        headers: {
                            "Content-Type": "application/json",
                            "Authorization": `Bearer ${accessToken}`,
                        },
                        body: JSON.stringify({
                            auth_id: user.sub,
                            name: user.name,
                            email: user.email,
                            account_type_id: "1",
                        }),
                    });
                    if (!postResponse.ok) {
                        const msg = await postResponse.text();
                        console.error("User registration failed:", msg);
                    }
                } else if (!getResponse.ok) {
                    const msg = await getResponse.text();
                    console.error("Error checking user existence:", msg);
                } else {
                    console.log("User already exists. Skipping registration.");
                }
            } catch (err) {
                console.error("Token or network error:", err);
            }
        };

        checkAndRegisterUser();
    }, [user.sub, accessToken]);

    const onDrop = useCallback((event: DragEvent<HTMLDivElement>) => {
        event.preventDefault();
        setIsDragging(false);
        setError(null);

        if (event.dataTransfer.files && event.dataTransfer.files.length > 0) {
            const file: File = event.dataTransfer.files[0];
            if (!allowedTypes.includes(file.type)) {
                setError("Chosen file type is not on the list of accepted types.");
                return;
            }
            setFile(file);
            setStartingType(file.type);
        }
    }, []);

    const onDragOver = useCallback((e: DragEvent<HTMLDivElement>) => {
        e.preventDefault();
        setIsDragging(true);
    }, []);

    const onDragLeave = useCallback((e: DragEvent<HTMLDivElement>) => {
        e.preventDefault();
        setIsDragging(false);
    }, []);

    const onFileChange = (e: ChangeEvent<HTMLInputElement>) => {
        setError(null);
        if (e.target.files && e.target.files.length > 0) {
            const file: File = e.target.files[0];
            if (!allowedTypes.includes(file.type)) {
                setError("Chosen file type is not on the list of accepted types.");
                return;
            }
            setStartingType(file.type);
            setFile(file);
        }
    };

    const onConvert = async () => {
        console.log(backendAddress)
        if (!file) return;
        if (startingType === targetType) {
            setError("File is already that type.");
            return;
        }

        setError(null);
        setIsConverting(true);

        try {
            const path: string = `${backendAddress}/${startingType.split("/").at(1)}/${targetType.split("/").at(1)}`;

            const response = await fetch(path, {
                method: "POST",
                headers: {
                    "Content-Type": startingType,
                    "Authorization": `Bearer ${accessToken}`,
                },
                body: new Blob([file], { type: file.type }),
            });

            if (!response.ok) {
                const text = await response.text();
                throw new Error(text || `Server returned ${response.status}`);
            }

            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement("a");
            a.href = url;
            a.download = file.name.replace(/\.[^.]+$/, "") + "." + targetType.split("/").at(1);
            document.body.appendChild(a);
            a.click();
            a.remove();
            window.URL.revokeObjectURL(url);
        } catch (err) {
            if (err instanceof Error) {
                setError(err.message);
            } else {
                setError("Conversion failed."); //needed to do it like that, eslint would stop
                //containerization otherwise
            }
        } finally {
            setIsConverting(false);
        }
    };

    const handleTargetTypeChosen = (e: React.ChangeEvent<HTMLSelectElement>) => {
        setTargetType(e.target.value);
    };

    return (
        <div style={styles.container}>
            <h1>Welcome, {user.name}!</h1>
            <button onClick={() => logout()}>
                Log out
            </button>
            <h1 style={styles.heading}>Image format converter</h1>

            <div>
                <label htmlFor="format">Select format:</label>
                <select id="format" value={targetType} onChange={handleTargetTypeChosen}>
                    <option value="image/jpeg">JPG</option>
                    <option value="image/png">PNG</option>
                    <option value="image/webp">WEBP</option>
                </select>
            </div>

            <div
                onDrop={onDrop}
                onDragOver={onDragOver}
                onDragLeave={onDragLeave}
                style={{
                    ...styles.dropArea,
                    ...(isDragging ? styles.dropAreaActive : {}),
                }}
            >
                {file ? (
                    <p>Selected file: <strong>{file.name}</strong></p>
                ) : (
                    <p>Drag & drop an image here, or click to select</p>
                )}
                <input
                    type="file"
                    accept="image/jpeg, image/png, image/webp"
                    style={styles.fileInput}
                    onChange={onFileChange}
                />
            </div>

            {error && <p style={styles.error}>{error}</p>}

            <button
                onClick={onConvert}
                disabled={!file || isConverting}
                style={{
                    ...styles.button,
                    ...((!file || isConverting) ? styles.buttonDisabled : {}),
                }}
            >
                {isConverting ? "Converting…" : "Convert"}
            </button>
        </div>
    );
}

const styles: { [key: string]: React.CSSProperties } = {
    container: {
        maxWidth: 480,
        margin: "4rem auto",
        padding: "1rem",
        textAlign: "center",
        fontFamily: "sans-serif",
    },
    heading: {
        marginBottom: "2rem",
    },
    dropArea: {
        position: "relative",
        border: "2px dashed #aaa",
        borderRadius: 8,
        padding: "2rem",
        cursor: "pointer",
        marginBottom: "1rem",
    },
    dropAreaActive: {
        border: "#333",
        backgroundColor: "#f5f5f5",
    },
    fileInput: {
        position: "absolute",
        top: 0,
        left: 0,
        width: "100%",
        height: "100%",
        opacity: 0,
        cursor: "pointer",
    },
    button: {
        padding: "0.75rem 1.5rem",
        fontSize: "1rem",
        borderRadius: 4,
        border: "none",
        backgroundColor: "#0070f3",
        color: "white",
        cursor: "pointer",
    },
    buttonDisabled: {
        backgroundColor: "#888",
        cursor: "not-allowed",
    },
    error: {
        color: "red",
        marginBottom: "1rem",
    },
};


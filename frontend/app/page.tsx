"use client";

import { useEffect, useState } from "react";

type Carrier = {
    id: string;
    name: string;
    transitDays: number;
    priceFactor: number;
};

type Quote = {
    origin: string;
    destination: string;
    weight: string;
    carrier: string;
    price: number;
    transitDays: number;
};

type Tracking = {
    trackingNumber: string;
    status: string;
    lastLocation: string;
    history: string[];
};

type ErrorResponse = {
    error: string;
};

export default function Home() {
    const [carriers, setCarriers] = useState<Carrier[]>([]);
    const [quote, setQuote] = useState<Quote | null>(null);
    const [tracking, setTracking] = useState<Tracking | null>(null);
    const [error, setError] = useState<string | null>(null);

    const backendUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

    useEffect(() => {
        async function loadData() {
            try {
                const carriersResp = await fetch(`${backendUrl}/api/carriers`);
                const carriersData = await carriersResp.json();
                setCarriers(carriersData);

                const quoteResp = await fetch(
                    `${backendUrl}/api/quote?origin=NYC&destination=LAX&weight=12`
                );
                const quoteData = await quoteResp.json();
                setQuote(quoteData);

                const trackingResp = await fetch(
                    `${backendUrl}/api/track?trackingNumber=TRACK123`
                );
                const trackingData = await trackingResp.json();
                setTracking(trackingData);
            } catch (err) {
                setError("Could not reach backend services. Make sure the Go services are running.");
            }
        }

        loadData();
    }, [backendUrl]);

    return (
        <main className="page-shell">
            <section className="hero">
                <h1>Multi-Carrier Shipping Dashboard</h1>
                <p>Example Next.js frontend calling a Go microservice backend gateway.</p>
            </section>

            {error ? <p className="error">{error}</p> : null}

            <section>
                <h2>Available carriers</h2>
                <div className="grid">
                    {carriers.length > 0 ? (
                        carriers.map((carrier) => (
                            <article key={carrier.id} className="card">
                                <h3>{carrier.name}</h3>
                                <p>Transit: {carrier.transitDays} days</p>
                                <p>Rate factor: {carrier.priceFactor}</p>
                            </article>
                        ))
                    ) : (
                        <p>Loading carriers…</p>
                    )}
                </div>
            </section>

            <section>
                <h2>Sample quote</h2>
                {quote ? (
                    <div className="card">
                        <p>
                            From <strong>{quote.origin}</strong> to <strong>{quote.destination}</strong>
                        </p>
                        <p>Weight: {quote.weight} kg</p>
                        <p>Carrier: {quote.carrier}</p>
                        <p>Price: ${quote.price.toFixed(2)}</p>
                        <p>Transit days: {quote.transitDays}</p>
                    </div>
                ) : (
                    <p>Loading quote…</p>
                )}
            </section>

            <section>
                <h2>Tracking status</h2>
                {tracking ? (
                    <div className="card">
                        <p>Tracking number: {tracking.trackingNumber}</p>
                        <p>Status: {tracking.status}</p>
                        <p>Last location: {tracking.lastLocation}</p>
                        <ul>
                            {tracking.history.map((event, index) => (
                                <li key={index}>{event}</li>
                            ))}
                        </ul>
                    </div>
                ) : (
                    <p>Loading tracking…</p>
                )}
            </section>
        </main>
    );
}

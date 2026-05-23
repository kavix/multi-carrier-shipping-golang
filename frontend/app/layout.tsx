import '../styles/globals.css';
import type { Metadata } from 'next';

export const metadata: Metadata = {
    title: 'Multi-Carrier Shipping',
    description: 'Shipping quotes, carriers, and tracking in a Go microservices backend',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
    return (
        <html lang="en">
            <body>{children}</body>
        </html>
    );
}

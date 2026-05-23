import './globals.css';
import type { Metadata } from 'next';

export const metadata: Metadata = {
    title: 'Antigravity Carrier Dashboard',
    description: 'Multi-carrier logistics dashboard powered by FreightPulse and MongoDB logs.',
};

export default function RootLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <html lang="en">
            <body>{children}</body>
        </html>
    );
}

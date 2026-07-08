import '@/styles/globals.css';
import type { Metadata } from 'next';
import { Inter } from 'next/font/google';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'Aegis SOC Dashboard',
  description: 'Adaptive Cyber Defense Platform',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="dark">
      <body className={`${inter.className} antialiased min-h-screen bg-background`}>
        {/* React Query Provider and Auth Provider wrappers will go here */}
        {children}
      </body>
    </html>
  );
}

export default function DashboardShell() {
  return (
    <main className="flex h-screen flex-col items-center justify-center">
      <div className="flex items-center gap-4">
        <div className="h-8 w-8 rounded-full bg-primary animate-pulse"></div>
        <h1 className="text-4xl font-bold tracking-tight">Aegis SOC</h1>
      </div>
      <p className="mt-4 text-muted-foreground max-w-lg text-center">
        Security Operations Center dashboard initialization complete. Awaiting module integration.
      </p>
    </main>
  );
}

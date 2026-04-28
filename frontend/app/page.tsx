import Link from "next/link";

export default function Home() {
  return (
    <main className="mx-auto flex max-w-3xl flex-col gap-6 p-8">
      <div>
        <h1 className="text-3xl font-semibold tracking-tight">Peer-to-peer car rental</h1>
        <p className="mt-2 text-slate-600">
          Owners list vehicles, customers discover cars, negotiate price over chat, and confirm
          bookings. The API is a Go/Gin service; this UI is Next.js.
        </p>
      </div>
      <div className="flex flex-wrap gap-3">
        <Link
          className="rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800"
          href="/register"
        >
          Create account
        </Link>
        <Link
          className="rounded-md border border-slate-300 px-4 py-2 text-sm font-medium hover:bg-slate-100"
          href="/login"
        >
          Sign in
        </Link>
        <Link
          className="rounded-md border border-slate-300 px-4 py-2 text-sm font-medium hover:bg-slate-100"
          href="/customer/search"
        >
          Browse cars
        </Link>
      </div>
      <p className="text-sm text-slate-500">
        Configure <code className="rounded bg-slate-200 px-1">NEXT_PUBLIC_API_URL</code> in{" "}
        <code className="rounded bg-slate-200 px-1">frontend/.env.local</code> if your API is not on{" "}
        <code className="rounded bg-slate-200 px-1">localhost:8080/api</code>.
      </p>
    </main>
  );
}

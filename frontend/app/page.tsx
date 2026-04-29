import type { SVGProps } from "react";
import Link from "next/link";

function IconSearch(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.75} aria-hidden {...props}>
      <circle cx="11" cy="11" r="7" />
      <path d="M20 20l-4.3-4.3" strokeLinecap="round" />
    </svg>
  );
}

function IconChat(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.75} aria-hidden {...props}>
      <path
        d="M8 10h.01M12 10h.01M16 10h.01M7 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 3v-3z"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function IconShield(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.75} aria-hidden {...props}>
      <path d="M12 3l8 4v6c0 5-3.5 8.5-8 9-4.5-.5-8-4-8-9V7l8-4z" strokeLinejoin="round" />
      <path d="M9 12l2 2 4-4" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function IconCar(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.75} aria-hidden {...props}>
      <path d="M7 17h.01M17 17h.01" strokeLinecap="round" />
      <path
        d="M5 17H3v-5l2-5h9l3 5h2a1 1 0 011 1v4h-2m-4 0H5m0 0a2 2 0 104 0m8 0a2 2 0 104 0M9 7V5h6v2"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export default function Home() {
  return (
    <div className="overflow-x-hidden">
      <section className="relative border-b border-emerald-900/30 bg-gradient-to-br from-slate-950 via-emerald-950 to-slate-950 text-white">
        <div
          className="pointer-events-none absolute inset-0 bg-[length:48px_48px] bg-grid-slate opacity-[0.35]"
          style={{
            maskImage: "radial-gradient(ellipse 80% 70% at 50% 0%, black 20%, transparent)",
            WebkitMaskImage: "radial-gradient(ellipse 80% 70% at 50% 0%, black 20%, transparent)",
          }}
        />
        <div className="relative mx-auto max-w-6xl px-4 pb-20 pt-16 sm:px-6 sm:pb-24 sm:pt-20 lg:px-8 lg:pt-24">
          <p className="mb-4 inline-flex items-center rounded-full border border-emerald-400/25 bg-emerald-500/10 px-3 py-1 text-xs font-medium tracking-wide text-emerald-200/90">
            Peer-to-peer car rental
          </p>
          <h1 className="max-w-3xl text-4xl font-semibold leading-[1.1] tracking-tight sm:text-5xl lg:text-6xl">
            Rent cars from people nearby—not only from lots.
          </h1>
          <p className="mt-6 max-w-2xl text-lg leading-relaxed text-slate-300 sm:text-xl">
            List your vehicle, set your rates, and chat with renters. Customers search by area, negotiate
            fairly, and confirm when the price is right.
          </p>
          <div className="mt-10 flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center">
            <Link
              href="/register"
              className="inline-flex items-center justify-center rounded-lg bg-emerald-400 px-6 py-3 text-sm font-semibold text-slate-950 shadow-lg shadow-emerald-900/40 transition hover:bg-emerald-300"
            >
              Create account
            </Link>
            <Link
              href="/login"
              className="inline-flex items-center justify-center rounded-lg border border-white/20 bg-white/5 px-6 py-3 text-sm font-semibold text-white backdrop-blur transition hover:bg-white/10"
            >
              Sign in
            </Link>
            <Link
              href="/customer/search"
              className="inline-flex items-center justify-center text-sm font-medium text-emerald-200/90 underline-offset-4 hover:text-white hover:underline sm:ml-2"
            >
              Browse cars →
            </Link>
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-6xl px-4 py-16 sm:px-6 lg:px-8 lg:py-20">
        <div className="mx-auto max-w-2xl text-center">
          <h2 className="text-2xl font-semibold tracking-tight text-slate-900 sm:text-3xl">
            Everything you need for short trips and long weekends
          </h2>
          <p className="mt-3 text-slate-600">
            Built for owners who want control and renters who want clarity—without the dealership feel.
          </p>
        </div>
        <ul className="mt-12 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {[
            {
              title: "Search by location",
              body: "Find vehicles in the area you need—filter by model, price, and availability.",
              icon: IconSearch,
              accent: "text-emerald-600 bg-emerald-50 ring-emerald-100",
            },
            {
              title: "Negotiate in one thread",
              body: "Owner and renter share a booking chat; agree on a final price before you confirm.",
              icon: IconChat,
              accent: "text-teal-600 bg-teal-50 ring-teal-100",
            },
            {
              title: "KYC & roles",
              body: "Customers and owners complete verification flows so both sides know who they are dealing with.",
              icon: IconShield,
              accent: "text-cyan-600 bg-cyan-50 ring-cyan-100",
            },
          ].map((item) => (
            <li
              key={item.title}
              className="group relative rounded-2xl border border-slate-200/80 bg-white p-6 shadow-sm ring-1 ring-slate-100 transition hover:border-slate-300 hover:shadow-md"
            >
              <div
                className={`mb-4 inline-flex h-12 w-12 items-center justify-center rounded-xl ring-1 ${item.accent}`}
              >
                <item.icon className="h-6 w-6" />
              </div>
              <h3 className="text-lg font-semibold text-slate-900">{item.title}</h3>
              <p className="mt-2 text-sm leading-relaxed text-slate-600">{item.body}</p>
            </li>
          ))}
        </ul>
      </section>

      <section className="border-y border-slate-200 bg-white py-16 sm:py-20">
        <div className="mx-auto max-w-6xl px-4 sm:px-6 lg:px-8">
          <h2 className="text-center text-2xl font-semibold tracking-tight text-slate-900 sm:text-3xl">
            Two sides, one platform
          </h2>
          <div className="mt-12 grid gap-8 lg:grid-cols-2">
            <div className="rounded-2xl border border-slate-200 bg-slate-50/80 p-5 sm:p-8">
              <div className="mb-4 flex h-11 w-11 items-center justify-center rounded-lg bg-slate-900 text-emerald-400">
                <IconCar className="h-6 w-6" />
              </div>
              <h3 className="text-xl font-semibold text-slate-900">Owners</h3>
              <p className="mt-2 text-slate-600">
                Add cars to your fleet, upload photos, set hourly, daily, and per-km pricing. Respond to
                inquiries, set the final booking price, and manage requests from one place.
              </p>
              <Link
                href="/register"
                className="mt-6 inline-flex text-sm font-semibold text-emerald-700 hover:text-emerald-800"
              >
                Register as owner →
              </Link>
            </div>
            <div className="rounded-2xl border border-emerald-200/80 bg-gradient-to-br from-emerald-50/90 to-white p-5 ring-1 ring-emerald-100 sm:p-8">
              <div className="mb-4 flex h-11 w-11 items-center justify-center rounded-lg bg-emerald-600 text-white">
                <IconSearch className="h-6 w-6" />
              </div>
              <h3 className="text-xl font-semibold text-slate-900">Customers</h3>
              <p className="mt-2 text-slate-600">
                Search cars near you, start a booking inquiry, and keep the conversation in-app until you
                confirm. Add your driving license on your account when you are ready to book.
              </p>
              <Link
                href="/customer/search"
                className="mt-6 inline-flex text-sm font-semibold text-emerald-700 hover:text-emerald-800"
              >
                Start searching →
              </Link>
            </div>
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-6xl px-4 py-16 sm:px-6 lg:px-8 lg:py-20">
        <h2 className="text-center text-2xl font-semibold tracking-tight text-slate-900 sm:text-3xl">
          How it works
        </h2>
        <ol className="mx-auto mt-12 grid max-w-4xl gap-8 sm:grid-cols-3">
          {[
            { step: "1", title: "Sign up", body: "Choose customer or owner and complete your profile." },
            { step: "2", title: "Connect", body: "Search or list; chat on the booking when there is interest." },
            { step: "3", title: "Confirm", body: "Agree on price, confirm the booking, and get on the road." },
          ].map((s) => (
            <li key={s.step} className="relative text-center">
              <span className="mx-auto flex h-10 w-10 items-center justify-center rounded-full bg-slate-900 text-sm font-bold text-emerald-400">
                {s.step}
              </span>
              <h3 className="mt-4 font-semibold text-slate-900">{s.title}</h3>
              <p className="mt-2 text-sm text-slate-600">{s.body}</p>
            </li>
          ))}
        </ol>
      </section>

      <section className="bg-slate-900 px-4 py-14 text-center text-white sm:px-6 lg:px-8">
        <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">Ready to try CarManage?</h2>
        <p className="mx-auto mt-3 max-w-lg text-slate-400">
          Create an account in seconds. Use the same app whether you are renting out a car or booking one.
        </p>
        <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row">
          <Link
            href="/register"
            className="inline-flex w-full max-w-xs items-center justify-center rounded-lg bg-emerald-400 px-6 py-3 text-sm font-semibold text-slate-950 hover:bg-emerald-300 sm:w-auto"
          >
            Get started
          </Link>
          <Link
            href="/login"
            className="inline-flex w-full max-w-xs items-center justify-center rounded-lg border border-white/25 px-6 py-3 text-sm font-semibold text-white hover:bg-white/10 sm:w-auto"
          >
            I already have an account
          </Link>
        </div>
      </section>

      <footer className="border-t border-slate-200 bg-slate-50 px-4 py-8 pb-[max(2rem,env(safe-area-inset-bottom))] sm:px-6 lg:px-8">
        <div className="mx-auto max-w-6xl text-center text-xs text-slate-500">
          <p className="break-words">
            Point the app at your API with{" "}
            <code className="rounded bg-slate-200/80 px-1.5 py-0.5 font-mono text-slate-700">
              NEXT_PUBLIC_API_URL
            </code>{" "}
            in{" "}
            <code className="rounded bg-slate-200/80 px-1.5 py-0.5 font-mono text-slate-700">
              frontend/.env.local
            </code>{" "}
            (default <code className="font-mono text-slate-700">http://localhost:8080/api</code>).
          </p>
          <p className="mt-2 text-slate-400">CarManage · P2P car rental</p>
        </div>
      </footer>
    </div>
  );
}

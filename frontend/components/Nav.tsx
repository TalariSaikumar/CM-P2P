"use client";

import type { SVGProps } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { clearSession, getToken, getUser, type User } from "@/lib/session";
import { useEffect, useState } from "react";

function IconMenu(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} aria-hidden {...props}>
      <path d="M4 6h16M4 12h16M4 18h16" strokeLinecap="round" />
    </svg>
  );
}

function IconClose(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} aria-hidden {...props}>
      <path d="M6 6l12 12M18 6L6 18" strokeLinecap="round" />
    </svg>
  );
}

export function Nav() {
  const pathname = usePathname();
  const router = useRouter();
  const [ready, setReady] = useState(false);
  const [logged, setLogged] = useState(false);
  const [user, setUser] = useState<User | null>(null);
  const [menuOpen, setMenuOpen] = useState(false);

  useEffect(() => {
    setLogged(!!getToken());
    setUser(getUser());
    setReady(true);
    setMenuOpen(false);
  }, [pathname]);

  function logout() {
    clearSession();
    setLogged(false);
    setUser(null);
    setMenuOpen(false);
    router.push("/");
    router.refresh();
  }

  const linkClass = "rounded-md py-2.5 text-sm text-slate-700 hover:bg-slate-50 hover:text-slate-900 md:inline md:py-0 md:hover:bg-transparent";
  const linkClassDesktop = "hover:text-slate-900";

  return (
    <header className="sticky top-0 z-40 border-b border-slate-200 bg-white/95 pt-[env(safe-area-inset-top)] backdrop-blur supports-[backdrop-filter]:bg-white/80">
      <div className="mx-auto flex max-w-6xl items-center justify-between gap-3 px-4 py-3 sm:px-6 lg:px-8">
        <Link href="/" className="shrink-0 font-semibold text-slate-900">
          CarManage
        </Link>
        <button
          type="button"
          className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-md text-slate-700 hover:bg-slate-100 md:hidden"
          aria-expanded={menuOpen}
          aria-controls="mobile-nav"
          aria-label={menuOpen ? "Close menu" : "Open menu"}
          onClick={() => setMenuOpen((o) => !o)}
        >
          {menuOpen ? <IconClose className="h-5 w-5" /> : <IconMenu className="h-5 w-5" />}
        </button>
        <nav className="hidden flex-wrap items-center gap-x-3 gap-y-2 text-sm text-slate-700 md:flex lg:gap-x-4">
          <Link className={linkClassDesktop} href="/customer/search">
            Search cars
          </Link>
          {ready && logged && user?.role === "CUSTOMER" && (
            <>
              <Link className={linkClassDesktop} href="/customer/bookings">
                My bookings
              </Link>
              <Link className={linkClassDesktop} href="/account">
                Account
              </Link>
            </>
          )}
          {ready && logged && user?.role === "OWNER" && (
            <>
              <Link className={linkClassDesktop} href="/owner/fleet">
                My fleet
              </Link>
              <Link className={linkClassDesktop} href="/owner/bookings">
                Booking requests
              </Link>
              <Link className={linkClassDesktop} href="/account">
                Account
              </Link>
            </>
          )}
          {ready && !logged && (
            <>
              <Link className={linkClassDesktop} href="/login">
                Sign in
              </Link>
              <Link
                className="rounded-md bg-slate-900 px-3 py-1.5 text-white hover:bg-slate-800"
                href="/register"
              >
                Register
              </Link>
            </>
          )}
          {ready && logged && (
            <button
              type="button"
              onClick={logout}
              className="rounded-md border border-slate-300 px-3 py-1.5 hover:bg-slate-50"
            >
              Sign out
            </button>
          )}
        </nav>
      </div>
      {menuOpen && (
        <div
          id="mobile-nav"
          className="max-h-[min(70vh,calc(100dvh-4rem))] overflow-y-auto border-t border-slate-200 bg-white px-4 pb-[max(1rem,env(safe-area-inset-bottom))] pt-2 md:hidden"
        >
          <div className="mx-auto flex max-w-6xl flex-col gap-0.5 sm:px-2">
            <Link className={`${linkClass} px-3`} href="/customer/search" onClick={() => setMenuOpen(false)}>
              Search cars
            </Link>
            {ready && logged && user?.role === "CUSTOMER" && (
              <>
                <Link className={`${linkClass} px-3`} href="/customer/bookings" onClick={() => setMenuOpen(false)}>
                  My bookings
                </Link>
                <Link className={`${linkClass} px-3`} href="/account" onClick={() => setMenuOpen(false)}>
                  Account
                </Link>
              </>
            )}
            {ready && logged && user?.role === "OWNER" && (
              <>
                <Link className={`${linkClass} px-3`} href="/owner/fleet" onClick={() => setMenuOpen(false)}>
                  My fleet
                </Link>
                <Link className={`${linkClass} px-3`} href="/owner/bookings" onClick={() => setMenuOpen(false)}>
                  Booking requests
                </Link>
                <Link className={`${linkClass} px-3`} href="/account" onClick={() => setMenuOpen(false)}>
                  Account
                </Link>
              </>
            )}
            {ready && !logged && (
              <>
                <Link className={`${linkClass} px-3`} href="/login" onClick={() => setMenuOpen(false)}>
                  Sign in
                </Link>
                <Link
                  className="mt-1 rounded-md bg-slate-900 px-3 py-3 text-center text-sm font-medium text-white hover:bg-slate-800"
                  href="/register"
                  onClick={() => setMenuOpen(false)}
                >
                  Register
                </Link>
              </>
            )}
            {ready && logged && (
              <button
                type="button"
                onClick={logout}
                className="mt-2 rounded-md border border-slate-300 px-3 py-3 text-left text-sm hover:bg-slate-50"
              >
                Sign out
              </button>
            )}
          </div>
        </div>
      )}
    </header>
  );
}

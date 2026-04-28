"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { clearSession, getToken, getUser, type User } from "@/lib/session";
import { useEffect, useState } from "react";

export function Nav() {
  const pathname = usePathname();
  const router = useRouter();
  const [ready, setReady] = useState(false);
  const [logged, setLogged] = useState(false);
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    setLogged(!!getToken());
    setUser(getUser());
    setReady(true);
  }, [pathname]);

  function logout() {
    clearSession();
    setLogged(false);
    setUser(null);
    router.push("/");
    router.refresh();
  }

  return (
    <header className="border-b border-slate-200 bg-white">
      <div className="mx-auto flex max-w-5xl items-center justify-between gap-4 px-4 py-3">
        <Link href="/" className="font-semibold text-slate-900">
          CarManage
        </Link>
        <nav className="flex flex-wrap items-center gap-3 text-sm text-slate-700">
          <Link className="hover:text-slate-900" href="/customer/search">
            Search cars
          </Link>
          {ready && logged && user?.role === "CUSTOMER" && (
            <>
              <Link className="hover:text-slate-900" href="/customer/bookings">
                My bookings
              </Link>
              <Link className="hover:text-slate-900" href="/account">
                Account
              </Link>
            </>
          )}
          {ready && logged && user?.role === "OWNER" && (
            <>
              <Link className="hover:text-slate-900" href="/owner/fleet">
                My fleet
              </Link>
              <Link className="hover:text-slate-900" href="/owner/bookings">
                Booking requests
              </Link>
              <Link className="hover:text-slate-900" href="/account">
                Account
              </Link>
            </>
          )}
          {ready && !logged && (
            <>
              <Link className="hover:text-slate-900" href="/login">
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
    </header>
  );
}

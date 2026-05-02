import { PageLoader } from "@/components/loaders";

/** Shown by Next.js while a route segment is loading (navigation / suspense). */
export default function AppLoading() {
  return (
    <div className="page-shell flex min-h-[50dvh] items-center justify-center">
      <PageLoader title="Loading…" subtitle="Getting things road-ready." className="min-h-0 py-4" />
    </div>
  );
}

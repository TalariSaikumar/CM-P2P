"use client";

type PaginationBarProps = {
  page: number;
  perPage: number;
  total: number;
  onPageChange: (page: number) => void;
  /** Plural noun for the count line, e.g. "vehicles" or "bookings" */
  noun: string;
};

export function PaginationBar({ page, perPage, total, onPageChange, noun }: PaginationBarProps) {
  if (total <= 0) return null;
  const totalPages = Math.max(1, Math.ceil(total / perPage));
  const from = (page - 1) * perPage + 1;
  const to = Math.min(page * perPage, total);

  return (
    <div className="flex flex-col gap-3 border-t border-slate-200 pt-4 sm:flex-row sm:items-center sm:justify-between">
      <p className="text-sm text-slate-600">
        {from}–{to} of {total} {noun}
        {totalPages > 1 && (
          <span className="text-slate-500">
            {" "}
            · Page {page} of {totalPages}
          </span>
        )}
      </p>
      {totalPages > 1 && (
        <div className="flex flex-wrap gap-2">
          <button
            type="button"
            disabled={page <= 1}
            onClick={() => onPageChange(page - 1)}
            className="min-h-[40px] rounded-md border border-slate-300 bg-white px-3 py-1.5 text-sm hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
          >
            Previous
          </button>
          <button
            type="button"
            disabled={page >= totalPages}
            onClick={() => onPageChange(page + 1)}
            className="min-h-[40px] rounded-md border border-slate-300 bg-white px-3 py-1.5 text-sm hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
          >
            Next
          </button>
        </div>
      )}
    </div>
  );
}

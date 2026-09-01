import Link from "next/link";

const links = [
  { href: "/upload", label: "Upload" },
  { href: "/videos", label: "Videos" },
];

export function Nav() {
  return (
    <header className="border-b border-surface-border bg-surface-raised/80 backdrop-blur">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
        <Link href="/" className="text-lg font-semibold tracking-tight text-white">
          Stream<span className="text-accent">Forge</span>
        </Link>
        <nav className="flex items-center gap-6 text-sm text-zinc-300">
          {links.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              className="transition hover:text-white"
            >
              {link.label}
            </Link>
          ))}
        </nav>
      </div>
    </header>
  );
}

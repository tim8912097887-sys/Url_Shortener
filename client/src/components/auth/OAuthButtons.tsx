function GoogleIcon(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg
      viewBox="0 0 48 48"
      width="18"
      height="18"
      aria-hidden="true"
      {...props}
    >
      <path
        fill="#FFC107"
        d="M43.6 20.5H42V20H24v8h11.3C33.7 32.9 29.3 36 24 36c-6.6 0-12-5.4-12-12s5.4-12 12-12c3.1 0 5.8 1.1 8 3l6-6C34.9 5.1 29.7 3 24 3 12.4 3 3 12.4 3 24s9.4 21 21 21 21-9.4 21-21c0-1.2-.1-2.4-.4-3.5z"
      />
      <path
        fill="#FF3D00"
        d="M6.3 14.7l6.6 4.8C14.6 15.9 18.9 13 24 13c3.1 0 5.8 1.1 8 3l6-6C34.9 5.1 29.7 3 24 3 16 3 9 7.5 6.3 14.7z"
      />
      <path
        fill="#4CAF50"
        d="M24 45c5.6 0 10.7-2.1 14.5-5.6l-6.7-5.5C29.7 35.7 27 36.5 24 36.5c-5.3 0-9.7-3.1-11.3-7.5l-6.6 5.1C9 40.5 16 45 24 45z"
      />
      <path
        fill="#1976D2"
        d="M43.6 20.5H42V20H24v8h11.3c-1 2.9-3.2 5.2-6 6.6l6.7 5.5C39.8 37.4 43 31.4 43 24c0-1.2-.1-2.4-.4-3.5z"
      />
    </svg>
  );
}

type OAuthButtonsProps = {
  href: string;
};

export default function OAuthButtons({ href }: OAuthButtonsProps) {
  return (
    <a
      href={href}
      className="flex w-full items-center justify-center gap-2 rounded-md border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 shadow-sm hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-teal-500 focus:ring-offset-2"
    >
      <GoogleIcon />
      Continue with Google
    </a>
  );
}

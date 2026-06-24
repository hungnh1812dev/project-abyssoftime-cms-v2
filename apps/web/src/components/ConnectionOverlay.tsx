interface ConnectionOverlayProps {
  visible: boolean;
}

export function ConnectionOverlay({ visible }: ConnectionOverlayProps) {
  return (
    <div
      role="alert"
      aria-live="assertive"
      aria-busy={visible}
      className={`fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm transition-opacity duration-300 ${
        visible ? 'opacity-100' : 'pointer-events-none opacity-0'
      }`}>
      <div className="flex flex-col items-center gap-4 text-center">
        <svg
          className="text-primary h-10 w-10 animate-spin"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          aria-hidden="true">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
        <div>
          <p className="text-foreground text-lg font-semibold">Connecting to service...</p>
          <p className="text-muted-foreground mt-1 text-sm">
            The server may be starting up. This can take up to 30 seconds.
          </p>
        </div>
      </div>
    </div>
  );
}

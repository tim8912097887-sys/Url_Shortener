type ReturnType<T, E> =
  | { ok: true; data: T; error: null }
  | { ok: false; data: null; error: E };
export const fetchWrapper = async <T, E>(
  fn: () => Promise<Response>,
): Promise<ReturnType<T, E>> => {
  const result = await fn();
  if (!result.ok) {
    return {
      ok: false,
      data: null,
      error: await result.json(),
    };
  }
  return {
    ok: true,
    data: await result.json(),
    error: null,
  };
};

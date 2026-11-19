export function attendanceLaunchToken(search: string, origin: string, key: "leave" | "launch") {
  try {
    const params = new URLSearchParams(search),
      nested = params.get("liff.state"),
      token =
        params.get(key) ||
        (nested ? new URL(nested, origin).searchParams.get(key) : "");
    return token && /^[a-f0-9]{64}$/.test(token) ? token : "";
  } catch {
    return "";
  }
}

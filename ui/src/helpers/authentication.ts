export function saveTokenCookie(token: string) {
  const maxAge = 14 * 24 * 60 * 60

  document.cookie = `auth-token=${token} ;max-age=${maxAge} ;path=/ ;sameSite=Lax`
}

export function clearTokenCookie() {
  document.cookie = 'auth-token= ;max-age=0 ;path=/ ;sameSite=Lax'
}

export function authToken() {
  const match = document.cookie.match(/auth-token=([\d\w]+)/)
  return match && match[1]
}

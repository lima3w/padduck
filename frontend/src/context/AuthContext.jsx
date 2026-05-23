import { createContext, useContext, useState, useEffect } from 'react'
import * as client from '../api/client'
import { clearCachedUser, getCachedUser, setCachedUser } from '../utils/storageKeys'

const AuthContext = createContext(null)

/** Provides shared authentication state to the whole app. */
export function AuthProvider({ children }) {
  const [user, setUserState] = useState(() => getCachedUser())
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  // Verify the session on mount and keep the cached user up to date.
  useEffect(() => {
    client.getCurrentUser()
      .then((res) => {
        setUserState(res.data)
        setCachedUser(res.data)
      })
      .catch(() => {
        setUserState(null)
        clearCachedUser()
      })
      .finally(() => setLoading(false))
  }, [])

  /** Set the current user and persist to cache. Pass null to clear. */
  const setUser = (userData) => {
    setUserState(userData)
    if (userData) setCachedUser(userData)
    else clearCachedUser()
  }

  const login = (userData) => {
    setUserState(userData)
    if (userData) setCachedUser(userData)
  }

  const logout = async () => {
    try { await client.logout() } catch {}
    setUserState(null)
    clearCachedUser()
  }

  const generateToken = async (tokenName) => {
    try {
      setError(null)
      const response = await client.generateTokenForMe(tokenName)
      return response.data.token
    } catch (err) {
      const errorMsg = err.response?.data?.error || 'Failed to generate token'
      setError(errorMsg)
      throw new Error(errorMsg)
    }
  }

  const listTokens = async () => {
    try {
      setError(null)
      const response = await client.listMyTokens()
      return response.data
    } catch (err) {
      const errorMsg = err.response?.data?.error || 'Failed to list tokens'
      setError(errorMsg)
      throw new Error(errorMsg)
    }
  }

  const revokeToken = async (tokenId) => {
    try {
      setError(null)
      await client.revokeToken(tokenId)
    } catch (err) {
      const errorMsg = err.response?.data?.error || 'Failed to revoke token'
      setError(errorMsg)
      throw new Error(errorMsg)
    }
  }

  return (
    <AuthContext.Provider value={{
      user,
      loading,
      error,
      login,
      logout,
      setUser,
      generateToken,
      listTokens,
      revokeToken,
      isAuthenticated: !!user,
    }}>
      {children}
    </AuthContext.Provider>
  )
}

/** Returns the shared auth state. Must be called inside <AuthProvider>. */
export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}

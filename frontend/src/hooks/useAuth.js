import { useState, useEffect } from 'react'
import * as client from '../api/client'
import { clearCachedUser, getCachedUser, setCachedUser } from '../utils/storageKeys'

export function useAuth() {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  // On mount, verify the session cookie is still valid by calling /auth/me.
  // Use the cached current user for an optimistic initial render, then confirm with the server.
  useEffect(() => {
    const cached = getCachedUser()
    if (cached) setUser(cached)
    client.getCurrentUser()
      .then((res) => {
        setUser(res.data)
        setCachedUser(res.data)
      })
      .catch(() => {
        setUser(null)
        clearCachedUser()
      })
      .finally(() => setLoading(false))
  }, [])

  const login = (userData) => {
    setUser(userData)
    if (userData) setCachedUser(userData)
  }

  const logout = async () => {
    try {
      await client.logout()
    } catch {}
    setUser(null)
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

  return {
    user,
    loading,
    error,
    login,
    logout,
    generateToken,
    listTokens,
    revokeToken,
    isAuthenticated: !!user,
  }
}

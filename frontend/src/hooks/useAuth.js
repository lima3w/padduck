import { useState, useEffect } from 'react'
import * as client from '../api/client'

export function useAuth() {
  const [user, setUser] = useState(null)
  const [token, setToken] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  // Load token and user from localStorage on mount
  useEffect(() => {
    const storedToken = localStorage.getItem('auth_token')
    const storedUser = localStorage.getItem('current_user')

    if (storedToken) {
      setToken(storedToken)
      if (storedUser) {
        setUser(JSON.parse(storedUser))
      }
      // Try to fetch fresh user data
      fetchCurrentUser(storedToken)
    }
    setLoading(false)
  }, [])

  const fetchCurrentUser = async (authToken) => {
    try {
      const response = await client.getCurrentUser()
      const userData = response.data
      setUser(userData)
      localStorage.setItem('current_user', JSON.stringify(userData))
    } catch (err) {
      console.error('Failed to fetch current user:', err)
      logout()
    }
  }

  const login = (token, user) => {
    setToken(token)
    setUser(user)
    localStorage.setItem('auth_token', token)
    localStorage.setItem('current_user', JSON.stringify(user))
  }

  const logout = () => {
    setToken(null)
    setUser(null)
    localStorage.removeItem('auth_token')
    localStorage.removeItem('current_user')
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
    token,
    loading,
    error,
    login,
    logout,
    generateToken,
    listTokens,
    revokeToken,
    isAuthenticated: !!token,
  }
}

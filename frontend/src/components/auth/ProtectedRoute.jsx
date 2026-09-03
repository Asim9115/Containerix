import { Navigate, Outlet } from 'react-router-dom'
import { useEffect } from 'react'
import { useAppDispatch, useAppSelector } from '../../store/hooks'
import { fetchProfile } from '../../store/slices/authSlice'
import { PageLoader } from '../ui/Loading'

export function ProtectedRoute() {
  const dispatch = useAppDispatch()
  const { isAuthenticated, loading, user } = useAppSelector((s) => s.auth)

  useEffect(() => {
    if (isAuthenticated && !user) {
      dispatch(fetchProfile())
    }
  }, [dispatch, isAuthenticated, user])

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  if (isAuthenticated && !user && loading) {
    return <PageLoader />
  }

  return <Outlet />
}

export function PublicRoute() {
  const isAuthenticated = useAppSelector((s) => s.auth.isAuthenticated)

  if (isAuthenticated) {
    return <Navigate to="/" replace />
  }

  return <Outlet />
}

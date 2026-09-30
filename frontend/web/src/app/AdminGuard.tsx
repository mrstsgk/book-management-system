import { Navigate, useLocation } from 'react-router-dom'
import { AdminLayout } from '@/components/layouts/AdminLayout'
import { LoadingState } from '@/components/states/LoadingState'
import { useLogout } from '@/features/admin-login/hooks/useLogout'
import { useSession } from '@/features/admin-login/hooks/useSession'

// /admin/* の親。ログイン済みのときだけ AdminLayout（と子画面）を出す
export function AdminGuard() {
  const { status } = useSession()
  const { logout, failed } = useLogout()
  const { pathname } = useLocation()

  if (status === 'loading') return <LoadingState />
  if (status === 'unauthorized')
    return <Navigate to="/admin/login" replace state={{ from: pathname }} />
  return <AdminLayout onLogout={logout} logoutFailed={failed} />
}

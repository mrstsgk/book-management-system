import { Navigate, useLocation } from 'react-router-dom'
import { AdminLayout } from '@/components/layouts/AdminLayout'
import { LoadingState } from '@/components/states/LoadingState'
import { useLogoutAction } from '@/features/admin-login/hooks/useLogoutAction'
import { useSession } from '@/features/admin-login/hooks/useSession'
import { useRedirectOnUnauthorizedRead } from '@/hooks/useRedirectOnUnauthorizedRead'

// /* の親。ログイン済みのときだけ AdminLayout（と子画面）を出す
export function AdminGuard() {
  const { status } = useSession()
  const { logout, failed } = useLogoutAction()
  const { pathname } = useLocation()
  useRedirectOnUnauthorizedRead()

  if (status === 'loading') return <LoadingState />
  if (status === 'unauthorized')
    return <Navigate to="/login" replace state={{ from: pathname }} />
  return <AdminLayout onLogout={logout} logoutFailed={failed} />
}

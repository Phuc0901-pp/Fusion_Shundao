import { Suspense, lazy } from 'react';
import { Routes, Route, Navigate, Outlet, useLocation } from 'react-router-dom';
import { Toaster } from 'sonner';
import { Layout } from './components/layout/Layout';
import { useTokenRefresh } from './hooks/useTokenRefresh';

// Code Splitting: Each page is loaded lazily (only when navigated to).
const Overview = lazy(() => import('./pages/Overview').then((m) => ({ default: m.Overview })));
const Login = lazy(() => import('./pages/Login').then((m) => ({ default: m.Login })));
const LockdownPage = lazy(() => import('./pages/Lockdown').then((m) => ({ default: m.Lockdown })));

const PageLoader = () => (
  <div className="flex items-center justify-center h-screen bg-slate-900 text-white text-xl md:text-2xl font-semibold opacity-60 animate-pulse">
    Đang tải hệ thống Shundao...
  </div>
);

// Protected Route wrapper checks sessionStorage auth state
const ProtectedRoute = () => {
  const isAuth = sessionStorage.getItem('shundao_auth') === 'true';
  const location = useLocation();

  // Auto-refresh JWT every 10 minutes while logged in
  useTokenRefresh();

  if (!isAuth) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  // Render Layout and its children routes if authenticated
  return (
    <Layout>
      <Suspense fallback={<PageLoader />}>
        <Outlet />
      </Suspense>
    </Layout>
  );
};

function App() {
  return (
    <>
      <Toaster
        position="top-right"
        richColors
        closeButton
        toastOptions={{
          duration: 4000,
          style: { fontSize: '13px' },
        }}
      />
      <Suspense fallback={<PageLoader />}>
        <Routes>
          {/* Public Routes */}
          <Route path="/login" element={<Login />} />
          <Route path="/lockdown" element={<LockdownPage />} />

          {/* Protected Routes Wrapper */}
          <Route element={<ProtectedRoute />}>
            <Route path="/" element={<Navigate to="/overview" replace />} />
            <Route path="/overview" element={<Overview />} />
            {/* Any other protected pages go here */}
          </Route>

          {/* 404 Route */}
          <Route path="*" element={<Navigate to="/login" replace />} />
        </Routes>
      </Suspense>
    </>
  );
}

export default App;

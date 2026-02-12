import { Routes, Route, Navigate } from 'react-router-dom';
import { Toaster } from 'sonner';
import { Layout } from './components/layout/Layout';
import { Overview } from './pages/Overview';

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
      <Layout>
        <Routes>
          <Route path="/" element={<Navigate to="/overview" replace />} />
          <Route path="/overview" element={<Overview />} />
          <Route path="*" element={<div className="text-white p-10 text-center">404 - Page Not Found</div>} />
        </Routes>
      </Layout>
    </>
  );
}

export default App;

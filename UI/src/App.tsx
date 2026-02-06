import { Routes, Route, Navigate } from 'react-router-dom';
import { Layout } from './components/layout/Layout';
import { Overview } from './pages/Overview';


function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Navigate to="/overview" replace />} />
        <Route path="/overview" element={<Overview />} />

        {/* Future routes: /meters, etc. */}
        <Route path="*" element={<div className="text-white p-10 text-center">404 - Page Not Found</div>} />
      </Routes>
    </Layout>
  );
}

export default App;

import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import SectionsPage from './pages/SectionsPage'
import SubnetsPage from './pages/SubnetsPage'
import IPAddressesPage from './pages/IPAddressesPage'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Navigate to="/sections" replace />} />
          <Route path="sections" element={<SectionsPage />} />
          <Route path="sections/:sectionID/subnets" element={<SubnetsPage />} />
          <Route path="subnets/:subnetID/ip-addresses" element={<IPAddressesPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

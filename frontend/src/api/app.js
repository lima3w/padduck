// App-level info: features, public info, dashboard.
import { api, noAuthApi } from './client'

export const getDashboardSummary = () => api.get('/dashboard/summary')

export const getDashboardRecentActivity = () => api.get('/dashboard/recent-activity')

export const getFeatures = () => api.get('/features')

export const getPublicInfo = () => noAuthApi.get('/public-info')

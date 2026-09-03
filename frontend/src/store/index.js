import { configureStore } from '@reduxjs/toolkit'
import authReducer from './slices/authSlice'
import deploymentsReducer from './slices/deploymentsSlice'
import containersReducer from './slices/containersSlice'
import jobsReducer from './slices/jobsSlice'
import buildReducer from './slices/buildSlice'
import adminReducer from './slices/adminSlice'

export const store = configureStore({
  reducer: {
    auth: authReducer,
    deployments: deploymentsReducer,
    containers: containersReducer,
    jobs: jobsReducer,
    build: buildReducer,
    admin: adminReducer,
  },
})

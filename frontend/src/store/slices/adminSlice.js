import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import {
  adminApi,
  healthApi,
  getStoredAdminApiKey,
  setStoredAdminApiKey,
} from '../../api'
import { ensureArray } from '../../utils/array'

function unauthorized(err, rejectWithValue) {
  if (err.status === 401) {
    setStoredAdminApiKey(null)
    return rejectWithValue({ message: err.message, unauthorized: true })
  }
  return rejectWithValue({ message: err.message, unauthorized: false })
}

export const loginAdmin = createAsyncThunk(
  'admin/login',
  async (apiKey, { rejectWithValue }) => {
    try {
      const key = apiKey.trim()
      await adminApi.verify(key)
      setStoredAdminApiKey(key)
      return key
    } catch (err) {
      setStoredAdminApiKey(null)
      return rejectWithValue(err.message)
    }
  },
)

export const fetchCgroup = createAsyncThunk(
  'admin/fetchCgroup',
  async (_, { rejectWithValue }) => {
    try {
      return await adminApi.getCgroup()
    } catch (err) {
      return unauthorized(err, rejectWithValue)
    }
  },
)

export const destroyCgroup = createAsyncThunk(
  'admin/destroyCgroup',
  async (_, { rejectWithValue }) => {
    try {
      return await adminApi.destroyCgroup()
    } catch (err) {
      return unauthorized(err, rejectWithValue)
    }
  },
)

export const fetchPorts = createAsyncThunk(
  'admin/fetchPorts',
  async (_, { rejectWithValue }) => {
    try {
      return await adminApi.getPorts()
    } catch (err) {
      return unauthorized(err, rejectWithValue)
    }
  },
)

export const fetchHealth = createAsyncThunk(
  'admin/fetchHealth',
  async (_, { rejectWithValue }) => {
    try {
      const [health, ready] = await Promise.all([
        healthApi.health(),
        healthApi.ready(),
      ])
      return { health, ready }
    } catch (err) {
      return rejectWithValue(err.message)
    }
  },
)

function applyAuthError(state, payload) {
  const message = typeof payload === 'string' ? payload : payload?.message
  state.error = message || null
  if (payload?.unauthorized) {
    state.isAuthenticated = false
    state.apiKey = null
  }
}

const adminSlice = createSlice({
  name: 'admin',
  initialState: {
    isAuthenticated: !!getStoredAdminApiKey(),
    apiKey: getStoredAdminApiKey(),
    cgroup: null,
    ports: [],
    health: null,
    ready: null,
    loading: false,
    authLoading: false,
    error: null,
    cgroupActionLoading: false,
  },
  reducers: {
    logoutAdmin(state) {
      setStoredAdminApiKey(null)
      state.isAuthenticated = false
      state.apiKey = null
      state.cgroup = null
      state.ports = []
      state.health = null
      state.ready = null
      state.error = null
    },
    clearAdminError(state) {
      state.error = null
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(loginAdmin.pending, (state) => {
        state.authLoading = true
        state.error = null
      })
      .addCase(loginAdmin.fulfilled, (state, action) => {
        state.authLoading = false
        state.isAuthenticated = true
        state.apiKey = action.payload
      })
      .addCase(loginAdmin.rejected, (state, action) => {
        state.authLoading = false
        state.isAuthenticated = false
        state.apiKey = null
        state.error = action.payload
      })
      .addCase(fetchCgroup.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(fetchCgroup.fulfilled, (state, action) => {
        state.loading = false
        state.cgroup = action.payload
      })
      .addCase(fetchCgroup.rejected, (state, action) => {
        state.loading = false
        applyAuthError(state, action.payload)
      })
      .addCase(destroyCgroup.pending, (state) => {
        state.cgroupActionLoading = true
      })
      .addCase(destroyCgroup.fulfilled, (state) => {
        state.cgroupActionLoading = false
        state.cgroup = null
      })
      .addCase(destroyCgroup.rejected, (state, action) => {
        state.cgroupActionLoading = false
        applyAuthError(state, action.payload)
      })
      .addCase(fetchPorts.fulfilled, (state, action) => {
        state.ports = ensureArray(action.payload)
      })
      .addCase(fetchPorts.rejected, (state, action) => {
        applyAuthError(state, action.payload)
      })
      .addCase(fetchHealth.fulfilled, (state, action) => {
        state.health = action.payload.health
        state.ready = action.payload.ready
      })
  },
})

export const { logoutAdmin, clearAdminError } = adminSlice.actions
export default adminSlice.reducer

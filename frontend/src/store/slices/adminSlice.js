import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import { adminApi, healthApi } from '../../api'
import { ensureArray } from '../../utils/array'

export const fetchCgroup = createAsyncThunk(
  'admin/fetchCgroup',
  async (_, { rejectWithValue }) => {
    try {
      return await adminApi.getCgroup()
    } catch (err) {
      return rejectWithValue(err.message)
    }
  },
)

export const destroyCgroup = createAsyncThunk(
  'admin/destroyCgroup',
  async (_, { rejectWithValue }) => {
    try {
      return await adminApi.destroyCgroup()
    } catch (err) {
      return rejectWithValue(err.message)
    }
  },
)

export const fetchPorts = createAsyncThunk(
  'admin/fetchPorts',
  async (_, { rejectWithValue }) => {
    try {
      return await adminApi.getPorts()
    } catch (err) {
      return rejectWithValue(err.message)
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

const adminSlice = createSlice({
  name: 'admin',
  initialState: {
    cgroup: null,
    ports: [],
    health: null,
    ready: null,
    loading: false,
    error: null,
    cgroupActionLoading: false,
  },
  reducers: {
    clearAdminError(state) {
      state.error = null
    },
  },
  extraReducers: (builder) => {
    builder
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
        state.error = action.payload
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
        state.error = action.payload
      })
      .addCase(fetchPorts.fulfilled, (state, action) => {
        state.ports = ensureArray(action.payload)
      })
      .addCase(fetchHealth.fulfilled, (state, action) => {
        state.health = action.payload.health
        state.ready = action.payload.ready
      })
  },
})

export const { clearAdminError } = adminSlice.actions
export default adminSlice.reducer

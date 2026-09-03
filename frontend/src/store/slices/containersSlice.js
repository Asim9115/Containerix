import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import { containersApi } from '../../api'
import { ensureArray } from '../../utils/array'

export const fetchContainers = createAsyncThunk(
  'containers/fetchAll',
  async (_, { rejectWithValue }) => {
    try {
      return await containersApi.list()
    } catch (err) {
      return rejectWithValue(err.message)
    }
  },
)

export const fetchContainer = createAsyncThunk(
  'containers/fetchOne',
  async (id, { rejectWithValue }) => {
    try {
      return await containersApi.get(id)
    } catch (err) {
      return rejectWithValue(err.message)
    }
  },
)

export const stopContainer = createAsyncThunk(
  'containers/stop',
  async (id, { rejectWithValue }) => {
    try {
      await containersApi.stop(id)
      return id
    } catch (err) {
      return rejectWithValue(err.message)
    }
  },
)

export const stopAllContainers = createAsyncThunk(
  'containers/stopAll',
  async (_, { rejectWithValue }) => {
    try {
      return await containersApi.stopAll()
    } catch (err) {
      return rejectWithValue(err.message)
    }
  },
)

export const deleteContainer = createAsyncThunk(
  'containers/delete',
  async (id, { rejectWithValue }) => {
    try {
      await containersApi.delete(id)
      return id
    } catch (err) {
      return rejectWithValue(err.message)
    }
  },
)

const containersSlice = createSlice({
  name: 'containers',
  initialState: {
    items: [],
    current: null,
    loading: false,
    error: null,
    lastStopped: [],
  },
  reducers: {
    clearContainerError(state) {
      state.error = null
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchContainers.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(fetchContainers.fulfilled, (state, action) => {
        state.loading = false
        state.items = ensureArray(action.payload)
      })
      .addCase(fetchContainers.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload
      })
      .addCase(fetchContainer.fulfilled, (state, action) => {
        state.current = action.payload
      })
      .addCase(stopContainer.fulfilled, (state, action) => {
        const id = action.payload
        state.items = state.items.map((c) =>
          c.ContainerID === id ? { ...c, Status: 'stopped' } : c,
        )
        if (state.current?.ContainerID === id) {
          state.current = { ...state.current, Status: 'stopped' }
        }
      })
      .addCase(stopAllContainers.fulfilled, (state, action) => {
        state.lastStopped = action.payload?.stopped || []
        state.items = state.items.map((c) =>
          state.lastStopped.includes(c.ContainerID) ? { ...c, Status: 'stopped' } : c,
        )
      })
      .addCase(deleteContainer.fulfilled, (state, action) => {
        state.items = state.items.filter((c) => c.ContainerID !== action.payload)
      })
  },
})

export const { clearContainerError } = containersSlice.actions
export default containersSlice.reducer

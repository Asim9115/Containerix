import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import { deploymentsApi } from '../../api'
import { ensureArray } from '../../utils/array'

export const fetchDeployments = createAsyncThunk(
  'deployments/fetchAll',
  async (_, { rejectWithValue }) => {
    try {
      return await deploymentsApi.list()
    } catch (err) {
      return rejectWithValue(err.message)
    }
  },
)

export const fetchDeployment = createAsyncThunk(
  'deployments/fetchOne',
  async (id, { rejectWithValue }) => {
    try {
      return await deploymentsApi.get(id)
    } catch (err) {
      return rejectWithValue(err.message)
    }
  },
)

export const deleteDeployment = createAsyncThunk(
  'deployments/delete',
  async (id, { rejectWithValue }) => {
    try {
      await deploymentsApi.delete(id)
      return id
    } catch (err) {
      return rejectWithValue(err.message)
    }
  },
)

const deploymentsSlice = createSlice({
  name: 'deployments',
  initialState: {
    items: [],
    current: null,
    loading: false,
    deleting: false,
    error: null,
    listRequestId: null,
    detailRequestId: null,
  },
  reducers: {
    clearCurrentDeployment(state) {
      state.current = null
    },
    clearDeploymentError(state) {
      state.error = null
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchDeployments.pending, (state, action) => {
        state.loading = true
        state.error = null
        state.listRequestId = action.meta.requestId
      })
      .addCase(fetchDeployments.fulfilled, (state, action) => {
        // Ignore stale responses so a slow poll cannot resurrect a deleted row
        if (state.listRequestId !== action.meta.requestId) return
        state.loading = false
        state.items = ensureArray(action.payload)
      })
      .addCase(fetchDeployments.rejected, (state, action) => {
        if (state.listRequestId !== action.meta.requestId) return
        state.loading = false
        state.error = action.payload
      })
      .addCase(fetchDeployment.pending, (state, action) => {
        state.loading = true
        state.error = null
        state.detailRequestId = action.meta.requestId
      })
      .addCase(fetchDeployment.fulfilled, (state, action) => {
        if (state.detailRequestId !== action.meta.requestId) return
        state.loading = false
        state.current = action.payload
      })
      .addCase(fetchDeployment.rejected, (state, action) => {
        if (state.detailRequestId !== action.meta.requestId) return
        state.loading = false
        state.error = action.payload
      })
      .addCase(deleteDeployment.pending, (state) => {
        state.deleting = true
        state.error = null
      })
      .addCase(deleteDeployment.fulfilled, (state, action) => {
        state.deleting = false
        state.items = state.items.filter((d) => d.ID !== action.payload)
        if (state.current?.ID === action.payload) {
          state.current = null
        }
      })
      .addCase(deleteDeployment.rejected, (state, action) => {
        state.deleting = false
        state.error = action.payload
      })
  },
})

export const { clearCurrentDeployment, clearDeploymentError } = deploymentsSlice.actions
export default deploymentsSlice.reducer

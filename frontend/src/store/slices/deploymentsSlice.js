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
    error: null,
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
      .addCase(fetchDeployments.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(fetchDeployments.fulfilled, (state, action) => {
        state.loading = false
        state.items = ensureArray(action.payload)
      })
      .addCase(fetchDeployments.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload
      })
      .addCase(fetchDeployment.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(fetchDeployment.fulfilled, (state, action) => {
        state.loading = false
        state.current = action.payload
      })
      .addCase(fetchDeployment.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload
      })
      .addCase(deleteDeployment.fulfilled, (state, action) => {
        state.items = state.items.filter((d) => d.ID !== action.payload)
        if (state.current?.ID === action.payload) {
          state.current = null
        }
      })
  },
})

export const { clearCurrentDeployment, clearDeploymentError } = deploymentsSlice.actions
export default deploymentsSlice.reducer

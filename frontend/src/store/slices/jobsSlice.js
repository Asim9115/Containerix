import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import { jobsApi } from '../../api'
import { ensureArray } from '../../utils/array'

export const fetchJobs = createAsyncThunk(
  'jobs/fetchAll',
  async (_, { rejectWithValue }) => {
    try {
      return await jobsApi.list()
    } catch (err) {
      return rejectWithValue(err.message)
    }
  },
)

export const fetchJob = createAsyncThunk(
  'jobs/fetchOne',
  async (id, { rejectWithValue }) => {
    try {
      return await jobsApi.get(id)
    } catch (err) {
      return rejectWithValue(err.message)
    }
  },
)

const jobsSlice = createSlice({
  name: 'jobs',
  initialState: {
    items: [],
    current: null,
    loading: false,
    error: null,
  },
  reducers: {
    clearCurrentJob(state) {
      state.current = null
    },
    clearJobError(state) {
      state.error = null
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchJobs.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(fetchJobs.fulfilled, (state, action) => {
        state.loading = false
        state.items = ensureArray(action.payload)
      })
      .addCase(fetchJobs.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload
      })
      .addCase(fetchJob.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(fetchJob.fulfilled, (state, action) => {
        state.loading = false
        state.current = action.payload
      })
      .addCase(fetchJob.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload
      })
  },
})

export const { clearCurrentJob, clearJobError } = jobsSlice.actions
export default jobsSlice.reducer

import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import { buildApi } from '../../api'

export const triggerBuild = createAsyncThunk(
  'build/trigger',
  async (data, { rejectWithValue }) => {
    try {
      return await buildApi.trigger(data)
    } catch (err) {
      return rejectWithValue(err.message)
    }
  },
)

const buildSlice = createSlice({
  name: 'build',
  initialState: {
    loading: false,
    error: null,
    lastBuild: null,
  },
  reducers: {
    clearBuildError(state) {
      state.error = null
    },
    clearLastBuild(state) {
      state.lastBuild = null
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(triggerBuild.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(triggerBuild.fulfilled, (state, action) => {
        state.loading = false
        state.lastBuild = action.payload
      })
      .addCase(triggerBuild.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload
      })
  },
})

export const { clearBuildError, clearLastBuild } = buildSlice.actions
export default buildSlice.reducer

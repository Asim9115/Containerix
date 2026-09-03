import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import { authApi, getStoredApiKey, setStoredApiKey } from '../../api'

export const registerUser = createAsyncThunk(
  'auth/register',
  async (data, { rejectWithValue }) => {
    try {
      const result = await authApi.register(data)
      setStoredApiKey(result.api_key)
      return result
    } catch (err) {
      return rejectWithValue(err.message)
    }
  },
)

export const loginWithApiKey = createAsyncThunk(
  'auth/loginWithApiKey',
  async (apiKey, { rejectWithValue }) => {
    try {
      setStoredApiKey(apiKey)
      const profile = await authApi.getProfile()
      return { profile, apiKey }
    } catch (err) {
      setStoredApiKey(null)
      return rejectWithValue(err.message)
    }
  },
)

export const fetchProfile = createAsyncThunk(
  'auth/fetchProfile',
  async (_, { rejectWithValue }) => {
    try {
      return await authApi.getProfile()
    } catch (err) {
      setStoredApiKey(null)
      return rejectWithValue(err.message)
    }
  },
)

export const rotateApiKey = createAsyncThunk(
  'auth/rotateApiKey',
  async (_, { rejectWithValue }) => {
    try {
      const result = await authApi.rotateApiKey()
      setStoredApiKey(result.api_key)
      return result
    } catch (err) {
      return rejectWithValue(err.message)
    }
  },
)

const initialState = {
  user: null,
  apiKey: getStoredApiKey(),
  isAuthenticated: !!getStoredApiKey(),
  loading: false,
  error: null,
  newApiKey: null,
}

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    logout(state) {
      setStoredApiKey(null)
      state.user = null
      state.apiKey = null
      state.isAuthenticated = false
      state.newApiKey = null
      state.error = null
    },
    clearNewApiKey(state) {
      state.newApiKey = null
    },
    clearError(state) {
      state.error = null
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(registerUser.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(registerUser.fulfilled, (state, action) => {
        state.loading = false
        state.isAuthenticated = true
        state.apiKey = action.payload.api_key
        state.newApiKey = action.payload.api_key
        state.user = {
          id: action.payload.user_id,
          email: action.payload.email,
          name: action.meta.arg.name,
        }
      })
      .addCase(registerUser.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload
      })
      .addCase(loginWithApiKey.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(loginWithApiKey.fulfilled, (state, action) => {
        state.loading = false
        state.isAuthenticated = true
        state.apiKey = action.payload.apiKey
        state.user = action.payload.profile
      })
      .addCase(loginWithApiKey.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload
      })
      .addCase(fetchProfile.pending, (state) => {
        state.loading = true
      })
      .addCase(fetchProfile.fulfilled, (state, action) => {
        state.loading = false
        state.isAuthenticated = true
        state.user = action.payload
      })
      .addCase(fetchProfile.rejected, (state) => {
        state.loading = false
        state.isAuthenticated = false
        state.user = null
        state.apiKey = null
      })
      .addCase(rotateApiKey.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(rotateApiKey.fulfilled, (state, action) => {
        state.loading = false
        state.apiKey = action.payload.api_key
        state.newApiKey = action.payload.api_key
      })
      .addCase(rotateApiKey.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload
      })
  },
})

export const { logout, clearNewApiKey, clearError } = authSlice.actions
export default authSlice.reducer

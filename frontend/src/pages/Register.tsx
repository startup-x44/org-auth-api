import React, { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import { Eye, EyeOff, Mail, Lock, User, AlertCircle, CheckCircle, Check, ArrowRight, Shield, Users } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import LoadingSpinner from '@/components/ui/loading-spinner'
import useAuthStore from '@/store/auth'
import { resolveTenant } from '@/utils/tenant'
import type { RegisterRequest } from '@/lib/types'

const Register = () => {
  const navigate = useNavigate()
  const { register, loading, error, isAuthenticated } = useAuthStore()

  const [formData, setFormData] = useState<RegisterRequest>({
    first_name: '',
    last_name: '',
    email: '',
    password: '',
    confirm_password: ''
  })
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const [localError, setLocalError] = useState('')
  const [passwordStrength, setPasswordStrength] = useState({
    length: false,
    uppercase: false,
    lowercase: false,
    number: false,
    special: false
  })
  const [currentStep, setCurrentStep] = useState(1)
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({})

  // Redirect if already authenticated
  useEffect(() => {
    if (isAuthenticated) {
      navigate('/dashboard', { replace: true })
    }
  }, [isAuthenticated, navigate])

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setFormData(prev => ({ ...prev, [name]: value }))

    if (name === 'password') {
      checkPasswordStrength(value)
    }

    // Clear validation error for this field
    if (validationErrors[name]) {
      setValidationErrors(prev => ({ ...prev, [name]: '' }))
    }

    if (localError) setLocalError('')
  }

  const checkPasswordStrength = (password: string) => {
    setPasswordStrength({
      length: password.length >= 8,
      uppercase: /[A-Z]/.test(password),
      lowercase: /[a-z]/.test(password),
      number: /\d/.test(password),
      special: /[!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?]/.test(password)
    })
  }

  const isPasswordValid = Object.values(passwordStrength).every(Boolean)
  const passwordStrengthScore = Object.values(passwordStrength).filter(Boolean).length
  const passwordStrengthPercentage = (passwordStrengthScore / 5) * 100

  const getPasswordStrengthColor = () => {
    if (passwordStrengthScore <= 2) return 'bg-red-500'
    if (passwordStrengthScore <= 3) return 'bg-yellow-500'
    if (passwordStrengthScore <= 4) return 'bg-blue-500'
    return 'bg-green-500'
  }

  const getPasswordStrengthLabel = () => {
    if (passwordStrengthScore <= 2) return 'Weak'
    if (passwordStrengthScore <= 3) return 'Fair'
    if (passwordStrengthScore <= 4) return 'Good'
    return 'Strong'
  }

  const validateForm = () => {
    const errors: Record<string, string> = {}

    if (!formData.first_name.trim()) errors.first_name = 'First name is required'
    if (!formData.last_name.trim()) errors.last_name = 'Last name is required'
    if (!formData.email.trim()) errors.email = 'Email is required'
    else if (!/\S+@\S+\.\S+/.test(formData.email)) errors.email = 'Please enter a valid email'
    if (!formData.password) errors.password = 'Password is required'
    else if (!isPasswordValid) errors.password = 'Password does not meet requirements'
    if (!formData.confirm_password) errors.confirm_password = 'Please confirm your password'
    else if (formData.password !== formData.confirm_password) errors.confirm_password = 'Passwords do not match'

    setValidationErrors(errors)
    return Object.keys(errors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLocalError('')

    if (!validateForm()) return

    // Resolve tenant from email domain
    const tenantId = resolveTenant(formData.email)
    if (!tenantId) {
      setLocalError('Unable to determine organization from email domain')
      return
    }

    const result = await register(formData)
    if (result.success) {
      navigate('/dashboard', { replace: true })
    } else {
      setLocalError(result.message || 'Registration failed')
    }
  }

  const togglePasswordVisibility = () => {
    setShowPassword(!showPassword)
  }

  const toggleConfirmPasswordVisibility = () => {
    setShowConfirmPassword(!showConfirmPassword)
  }

  const steps = [
    { id: 1, title: 'Personal Info', icon: User },
    { id: 2, title: 'Account Setup', icon: Mail },
    { id: 3, title: 'Security', icon: Shield }
  ]

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-100 dark:from-slate-800 dark:via-slate-700 dark:to-slate-600 p-4">
      {/* Background Pattern */}
      <div className="absolute inset-0 bg-grid-slate-200/50 [mask-image:linear-gradient(0deg,white,rgba(255,255,255,0.8))] dark:bg-grid-slate-700/30 dark:[mask-image:linear-gradient(0deg,rgba(255,255,255,0.05),rgba(255,255,255,0.2))]" />

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, ease: "easeOut" }}
        className="w-full max-w-2xl relative z-10 text-slate-800 dark:text-slate-100"
      >
        {/* Progress Steps */}
        <div className="mb-8">
          <div className="flex items-center justify-center space-x-4">
            {steps.map((step, index) => {
              const Icon = step.icon
              const isActive = currentStep >= step.id
              const isCompleted = currentStep > step.id

              return (
                <React.Fragment key={step.id}>
                  <motion.div
                    initial={{ scale: 0 }}
                    animate={{ scale: 1 }}
                    transition={{ delay: index * 0.1 }}
                    className="flex flex-col items-center"
                  >
                    <motion.div
                      className={`w-12 h-12 rounded-full flex items-center justify-center border-2 transition-all duration-300 ${
                        isCompleted
                          ? 'bg-emerald-600 border-emerald-600 text-white shadow-lg'
                          : isActive
                          ? 'bg-blue-600 border-blue-600 text-white shadow-lg'
                          : 'bg-white dark:bg-slate-800 border-slate-300 dark:border-slate-600 text-slate-400 dark:text-slate-500 shadow-sm'
                      }`}
                      whileHover={{ scale: 1.05 }}
                      whileTap={{ scale: 0.95 }}
                    >
                      {isCompleted ? (
                        <Check className="w-5 h-5" />
                      ) : (
                        <Icon className="w-5 h-5" />
                      )}
                    </motion.div>
                    <span className={`text-xs mt-2 font-medium ${
                      isActive ? 'text-blue-700 dark:text-blue-300' : 'text-slate-500 dark:text-slate-400'
                    }`}>
                      {step.title}
                    </span>
                  </motion.div>
                  {index < steps.length - 1 && (
                    <motion.div
                      className={`w-16 h-0.5 transition-colors duration-300 ${
                        currentStep > step.id ? 'bg-emerald-600' : 'bg-slate-300 dark:bg-slate-600'
                      }`}
                      initial={{ scaleX: 0 }}
                      animate={{ scaleX: 1 }}
                      transition={{ delay: (index + 1) * 0.1 }}
                    />
                  )}
                </React.Fragment>
              )
            })}
          </div>
        </div>

  <Card className="shadow-2xl border border-slate-200/50 dark:border-slate-600/50 bg-white/95 dark:bg-slate-800/88 backdrop-blur-md">
          <CardHeader className="space-y-1 text-center pb-8">
            <motion.div
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
            >
              <CardTitle className="text-3xl font-bold bg-gradient-to-r from-slate-900 via-blue-800 to-slate-900 dark:from-slate-100 dark:via-blue-200 dark:to-slate-100 bg-clip-text text-transparent">
                Create Your Account
              </CardTitle>
              <CardDescription className="text-lg text-slate-600 dark:text-slate-300 mt-2">
                Join thousands of organizations using our secure platform
              </CardDescription>
            </motion.div>
          </CardHeader>
          <CardContent className="px-8 pb-8">
            <form onSubmit={handleSubmit} className="space-y-6">
              {(error || localError) && (
                <motion.div
                  initial={{ opacity: 0, scale: 0.95 }}
                  animate={{ opacity: 1, scale: 1 }}
                  exit={{ opacity: 0, scale: 0.95 }}
                >
                  <Alert variant="destructive" className="border-red-200 bg-red-50 dark:bg-red-950/50">
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription className="font-medium">
                      {localError || error}
                    </AlertDescription>
                  </Alert>
                </motion.div>
              )}

              <AnimatePresence mode="wait">
                <motion.div
                  key={currentStep}
                  initial={{ opacity: 0, x: 20 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: -20 }}
                  transition={{ duration: 0.3 }}
                  className="space-y-6"
                >

                  {/* Step 1: Personal Information */}
                  {currentStep === 1 && (
                    <div className="space-y-6">
                      <div className="text-center mb-6">
                        <h3 className="text-xl font-semibold text-foreground">Personal Information</h3>
                        <p className="text-sm text-slate-600 dark:text-slate-300 mt-1">Tell us a bit about yourself</p>
                      </div>

                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className="space-y-2">
                          <Label htmlFor="first_name" className="text-sm font-medium text-slate-700 dark:text-slate-300">
                            First Name *
                          </Label>
                          <div className="relative">
                            <User className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                            <Input
                              id="first_name"
                              name="first_name"
                              type="text"
                              placeholder="John"
                              value={formData.first_name}
                              onChange={handleInputChange}
                              className={`pl-10 border-input dark:border-slate-600 focus:border-blue-500 dark:focus:border-blue-400 bg-background dark:bg-slate-800 text-foreground placeholder:text-muted-foreground transition-all duration-200 ${
                                validationErrors.first_name ? 'border-red-500 focus:border-red-500' : ''
                              }`}
                              disabled={loading}
                              required
                            />
                          </div>
                          {validationErrors.first_name && (
                            <p className="text-xs text-red-600 flex items-center gap-1">
                              <AlertCircle className="h-3 w-3" />
                              {validationErrors.first_name}
                            </p>
                          )}
                        </div>

                        <div className="space-y-2">
                          <Label htmlFor="last_name" className="text-sm font-medium text-slate-700 dark:text-slate-300">
                            Last Name *
                          </Label>
                          <div className="relative">
                            <User className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                            <Input
                              id="last_name"
                              name="last_name"
                              type="text"
                              placeholder="Doe"
                              value={formData.last_name}
                              onChange={handleInputChange}
                              className={`pl-10 border-slate-300 dark:border-slate-600 focus:border-blue-500 dark:focus:border-blue-400 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 placeholder:text-slate-500 dark:placeholder:text-slate-400 transition-all duration-200 ${
                                validationErrors.last_name ? 'border-red-500 focus:border-red-500' : ''
                              }`}
                              disabled={loading}
                              required
                            />
                          </div>
                          {validationErrors.last_name && (
                            <p className="text-xs text-red-600 flex items-center gap-1">
                              <AlertCircle className="h-3 w-3" />
                              {validationErrors.last_name}
                            </p>
                          )}
                        </div>
                      </div>
                    </div>
                  )}

                  {/* Step 2: Account Setup */}
                  {currentStep === 2 && (
                    <div className="space-y-6">
                      <div className="text-center mb-6">
                        <h3 className="text-xl font-semibold text-foreground">Account Setup</h3>
                        <p className="text-sm text-slate-600 dark:text-slate-300 mt-1">Set up your login credentials</p>
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="email" className="text-sm font-medium text-slate-700 dark:text-slate-300">
                          Work Email *
                        </Label>
                        <div className="relative">
                          <Mail className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                          <Input
                            id="email"
                            name="email"
                            type="email"
                            placeholder="john.doe@company.com"
                            value={formData.email}
                            onChange={handleInputChange}
                            className={`pl-10 border-input dark:border-slate-600 focus:border-blue-500 dark:focus:border-blue-400 bg-background dark:bg-slate-800 text-foreground placeholder:text-muted-foreground transition-all duration-200 ${
                              validationErrors.email ? 'border-red-500 focus:border-red-500' : ''
                            }`}
                            disabled={loading}
                            required
                          />
                        </div>
                        {validationErrors.email && (
                          <p className="text-xs text-red-600 flex items-center gap-1">
                            <AlertCircle className="h-3 w-3" />
                            {validationErrors.email}
                          </p>
                        )}
                        <p className="text-xs text-slate-600 dark:text-slate-300 flex items-center gap-1">
                          <Users className="h-3 w-3" />
                          Your organization will be automatically detected from your email domain
                        </p>
                      </div>
                    </div>
                  )}

                  {/* Step 3: Security */}
                  {currentStep === 3 && (
                    <div className="space-y-6">
                      <div className="text-center mb-6">
                        <h3 className="text-xl font-semibold text-foreground">Security</h3>
                        <p className="text-sm text-slate-600 dark:text-slate-300 mt-1">Create a strong password to protect your account</p>
                      </div>

                      <div className="space-y-4">
                        <div className="space-y-2">
                          <Label htmlFor="password" className="text-sm font-medium text-slate-700 dark:text-slate-300">
                            Password *
                          </Label>
                          <div className="relative">
                            <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                            <Input
                              id="password"
                              name="password"
                              type={showPassword ? 'text' : 'password'}
                              placeholder="Create a strong password"
                              value={formData.password}
                              onChange={handleInputChange}
                              className={`pl-10 pr-10 border-input dark:border-slate-600 focus:border-blue-500 dark:focus:border-blue-400 bg-background dark:bg-slate-800 text-foreground placeholder:text-muted-foreground transition-all duration-200 ${
                                validationErrors.password ? 'border-red-500 focus:border-red-500' : ''
                              }`}
                              disabled={loading}
                              required
                            />
                            <button
                              type="button"
                              onClick={togglePasswordVisibility}
                              className="absolute right-3 top-3 text-muted-foreground hover:text-foreground transition-colors"
                              disabled={loading}
                            >
                              {showPassword ? (
                                <EyeOff className="h-4 w-4" />
                              ) : (
                                <Eye className="h-4 w-4" />
                              )}
                            </button>
                          </div>

                          {/* Password Strength Indicator */}
                          {formData.password && (
                            <div className="space-y-3">
                              <div className="flex items-center justify-between">
                                <span className="text-xs font-medium text-slate-700 dark:text-slate-200">Password Strength</span>
                                <Badge variant="outline" className={`text-xs font-medium border-2 ${
                                  passwordStrengthScore <= 2 ? 'border-red-400 text-red-700 dark:border-red-500 dark:text-red-300 bg-red-50 dark:bg-red-950' :
                                  passwordStrengthScore <= 3 ? 'border-yellow-400 text-yellow-700 dark:border-yellow-500 dark:text-yellow-300 bg-yellow-50 dark:bg-yellow-950' :
                                  passwordStrengthScore <= 4 ? 'border-blue-400 text-blue-700 dark:border-blue-500 dark:text-blue-300 bg-blue-50 dark:bg-blue-950' :
                                  'border-emerald-400 text-emerald-700 dark:border-emerald-500 dark:text-emerald-300 bg-emerald-50 dark:bg-emerald-950'
                                }`}>
                                  {getPasswordStrengthLabel()}
                                </Badge>
                              </div>
                              <div className="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-2.5">
                                <motion.div
                                  className={`h-2.5 rounded-full ${getPasswordStrengthColor()}`}
                                  initial={{ width: 0 }}
                                  animate={{ width: `${passwordStrengthPercentage}%` }}
                                  transition={{ duration: 0.3 }}
                                />
                              </div>
                              <div className="grid grid-cols-1 gap-1 text-xs">
                                {[
                                  { key: 'length', label: 'At least 8 characters' },
                                  { key: 'uppercase', label: 'One uppercase letter' },
                                  { key: 'lowercase', label: 'One lowercase letter' },
                                  { key: 'number', label: 'One number' },
                                  { key: 'special', label: 'One special character' }
                                ].map(({ key, label }) => (
                                  <div key={key} className={`flex items-center gap-2 transition-colors ${
                                    passwordStrength[key as keyof typeof passwordStrength] ? 'text-emerald-700 dark:text-emerald-300' : 'text-slate-500 dark:text-slate-300'
                                  }`}>
                                    {passwordStrength[key as keyof typeof passwordStrength] ? (
                                      <CheckCircle className="h-3 w-3" />
                                    ) : (
                                      <div className="h-3 w-3 rounded-full border border-current" />
                                    )}
                                    {label}
                                  </div>
                                ))}
                              </div>
                            </div>
                          )}

                          {validationErrors.password && (
                            <p className="text-xs text-red-600 flex items-center gap-1">
                              <AlertCircle className="h-3 w-3" />
                              {validationErrors.password}
                            </p>
                          )}
                        </div>

                        <div className="space-y-2">
                          <Label htmlFor="confirm_password" className="text-sm font-medium text-slate-700 dark:text-slate-300">
                            Confirm Password *
                          </Label>
                          <div className="relative">
                            <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                            <Input
                              id="confirm_password"
                              name="confirm_password"
                              type={showConfirmPassword ? 'text' : 'password'}
                              placeholder="Confirm your password"
                              value={formData.confirm_password}
                              onChange={handleInputChange}
                              className={`pl-10 pr-10 border-input dark:border-slate-600 focus:border-blue-500 dark:focus:border-blue-400 bg-background dark:bg-slate-800 text-foreground placeholder:text-muted-foreground transition-all duration-200 ${
                                validationErrors.confirm_password ? 'border-red-500 focus:border-red-500' : ''
                              }`}
                              disabled={loading}
                              required
                            />
                            <button
                              type="button"
                              onClick={toggleConfirmPasswordVisibility}
                              className="absolute right-3 top-3 text-muted-foreground hover:text-foreground transition-colors"
                              disabled={loading}
                            >
                              {showConfirmPassword ? (
                                <EyeOff className="h-4 w-4" />
                              ) : (
                                <Eye className="h-4 w-4" />
                              )}
                            </button>
                          </div>
                          {validationErrors.confirm_password && (
                            <p className="text-xs text-red-600 flex items-center gap-1">
                              <AlertCircle className="h-3 w-3" />
                              {validationErrors.confirm_password}
                            </p>
                          )}
                        </div>
                      </div>
                    </div>
                  )}
                </motion.div>
              </AnimatePresence>

              {/* Navigation Buttons */}
              <div className="flex justify-between pt-6">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setCurrentStep(Math.max(1, currentStep - 1))}
                  disabled={currentStep === 1 || loading}
                  className="px-6"
                >
                  Previous
                </Button>

                {currentStep < 3 ? (
                  <Button
                    type="button"
                    onClick={() => {
                      if (currentStep === 1 && (!formData.first_name.trim() || !formData.last_name.trim())) {
                        setValidationErrors({
                          first_name: !formData.first_name.trim() ? 'First name is required' : '',
                          last_name: !formData.last_name.trim() ? 'Last name is required' : ''
                        })
                        return
                      }
                      if (currentStep === 2 && (!formData.email.trim() || !/\S+@\S+\.\S+/.test(formData.email))) {
                        setValidationErrors({
                          email: !formData.email.trim() ? 'Email is required' :
                                 !/\S+@\S+\.\S+/.test(formData.email) ? 'Please enter a valid email' : ''
                        })
                        return
                      }
                      setCurrentStep(currentStep + 1)
                    }}
                    disabled={loading}
                    className="px-6"
                  >
                    Next
                    <ArrowRight className="w-4 h-4 ml-2" />
                  </Button>
                ) : (
                  <Button
                    type="submit"
                    disabled={loading || !isPasswordValid}
                    className="px-8 bg-gradient-to-r from-blue-600 via-purple-600 to-blue-700 hover:from-blue-700 hover:via-purple-700 hover:to-blue-800 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {loading ? (
                      <>
                        <LoadingSpinner size="sm" className="mr-2" />
                        Creating Account...
                      </>
                    ) : (
                      <>
                        Create Account
                        <Check className="w-4 h-4 ml-2" />
                      </>
                    )}
                  </Button>
                )}
              </div>
            </form>

            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 0.5 }}
              className="mt-8 text-center"
            >
              <span className="text-slate-600 dark:text-slate-400">Already have an account? </span>
              <Link
                to="/login"
                className="text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 font-medium transition-colors"
              >
                Sign in here
              </Link>
            </motion.div>
          </CardContent>
        </Card>
      </motion.div>
    </div>
  )
}

export default Register
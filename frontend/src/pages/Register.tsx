import React, { useState, useEffect } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import { Eye, EyeOff, Mail, Lock, User, AlertCircle, CheckCircle, Check, ArrowRight, Shield, Users, Phone, UserPlus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import LoadingSpinner from '@/components/ui/loading-spinner'
import useAuthStore from '@/store/auth'
import type { RegisterRequest } from '@/lib/types'

const Register = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const invitationToken = searchParams.get('invitation_token')
  const invitationEmail = searchParams.get('email') // Email from invitation link
  const { register, loading, error, isAuthenticated, clearError } = useAuthStore()

  const [formData, setFormData] = useState<RegisterRequest>({
    first_name: '',
    last_name: '',
    email: invitationEmail || '', // Pre-fill email if from invitation
    phone: '',
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

  // Clear errors when component mounts
  useEffect(() => {
    if (clearError) clearError()
    setLocalError('')
  }, [])

  // Pre-fill email from invitation
  useEffect(() => {
    if (invitationEmail) {
      setFormData(prev => ({ ...prev, email: invitationEmail }))
    }
  }, [invitationEmail])

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
    if (error && clearError) clearError()
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
    // Allow + character and other valid email characters
    else if (!/^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/.test(formData.email)) {
      errors.email = 'Please enter a valid email'
    }
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

    // Include invitation token if present
    const registrationData = {
      ...formData,
      ...(invitationToken && { invitation_token: invitationToken })
    }

    const result = await register(registrationData)
    if (result.success) {
      // Redirect to email verification page
      navigate(`/verify-email?email=${encodeURIComponent(formData.email)}`, { replace: true })
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
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 via-purple-50 to-pink-50 dark:from-slate-900 dark:via-purple-900 dark:to-slate-900 p-4">
      {/* Animated Background Pattern */}
      <div className="absolute inset-0 bg-grid-slate-200/50 [mask-image:radial-gradient(ellipse_at_center,white,transparent)] dark:bg-grid-slate-700/30" />
      
      {/* Floating Orbs */}
      <motion.div
        animate={{
          scale: [1, 1.2, 1],
          opacity: [0.3, 0.5, 0.3],
        }}
        transition={{
          duration: 8,
          repeat: Infinity,
          ease: "easeInOut"
        }}
        className="absolute top-20 left-20 w-72 h-72 bg-purple-400/30 rounded-full blur-3xl"
      />
      <motion.div
        animate={{
          scale: [1, 1.3, 1],
          opacity: [0.2, 0.4, 0.2],
        }}
        transition={{
          duration: 10,
          repeat: Infinity,
          ease: "easeInOut",
          delay: 1
        }}
        className="absolute bottom-20 right-20 w-96 h-96 bg-pink-400/20 rounded-full blur-3xl"
      />

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, ease: "easeOut" }}
        className="w-full max-w-2xl relative z-10"
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
                      className={`w-14 h-14 rounded-2xl flex items-center justify-center border-2 transition-all duration-300 ${
                        isCompleted
                          ? 'bg-gradient-to-br from-green-500 to-emerald-600 border-green-500 text-white shadow-lg shadow-green-500/50'
                          : isActive
                          ? 'bg-gradient-to-br from-blue-500 to-purple-600 border-blue-500 text-white shadow-lg shadow-blue-500/50'
                          : 'bg-white/80 dark:bg-slate-800/80 backdrop-blur-sm border-slate-300 dark:border-slate-600 text-slate-400 dark:text-slate-500 shadow-sm'
                      }`}
                      whileHover={{ scale: 1.05 }}
                      whileTap={{ scale: 0.95 }}
                    >
                      {isCompleted ? (
                        <Check className="w-6 h-6" />
                      ) : (
                        <Icon className="w-6 h-6" />
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

        {/* Main Registration Card */}
        <Card className="shadow-2xl border-0 bg-white/90 dark:bg-slate-800/90 backdrop-blur-xl overflow-hidden">
          {/* Gradient Header */}
          <div className="relative overflow-hidden bg-gradient-to-br from-purple-600 via-indigo-600 to-blue-600 px-8 py-10">
            <div className="absolute top-0 right-0 -mt-12 -mr-12 w-48 h-48 rounded-full bg-white/10 backdrop-blur-sm"></div>
            <div className="absolute bottom-0 left-0 -mb-12 -ml-12 w-36 h-36 rounded-full bg-white/10 backdrop-blur-sm"></div>
            <motion.div
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
              className="relative z-10"
            >
              <div className="flex items-center justify-center mb-4">
                <div className="p-3 rounded-2xl bg-white/20 backdrop-blur-sm">
                  <UserPlus className="w-8 h-8 text-white" />
                </div>
              </div>
              <CardTitle className="text-3xl font-bold text-white text-center">
                {invitationToken ? 'Complete Your Registration' : 'Create Your Account'}
              </CardTitle>
              <CardDescription className="text-purple-100 text-center mt-2 text-base">
                {invitationToken 
                  ? 'You\'ve been invited to join an organization. Create your account to accept the invitation.'
                  : 'Join thousands of organizations using our secure platform'
                }
              </CardDescription>
            </motion.div>
          </div>
          
          <CardHeader className="sr-only">
            <CardTitle>Register</CardTitle>
          </CardHeader>
          <CardContent className="px-8 pb-8">
            {invitationToken && (
              <motion.div
                initial={{ opacity: 0, y: -10 }}
                animate={{ opacity: 1, y: 0 }}
                className="mb-6"
              >
                <Alert className="border-purple-200 bg-purple-50 dark:bg-purple-950/30">
                  <Mail className="h-4 w-4 text-purple-600" />
                  <AlertDescription className="text-purple-800 dark:text-purple-200">
                    <strong>Organization Invitation:</strong> After creating your account, you'll automatically join the organization that invited you.
                  </AlertDescription>
                </Alert>
              </motion.div>
            )}
            
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
                      <div className="text-center mb-8">
                        <h3 className="text-2xl font-bold bg-gradient-to-r from-purple-600 to-blue-600 bg-clip-text text-transparent">
                          Personal Information
                        </h3>
                        <p className="text-sm text-slate-600 dark:text-slate-400 mt-2">Tell us a bit about yourself</p>
                      </div>

                      <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                        <div className="space-y-2">
                          <Label htmlFor="first_name" className="text-base font-semibold text-slate-700 dark:text-slate-300">
                            First Name *
                          </Label>
                          <div className="relative group">
                            <User className="absolute left-4 top-4 h-5 w-5 text-slate-400 group-focus-within:text-purple-500 transition-colors" />
                            <Input
                              id="first_name"
                              name="first_name"
                              type="text"
                              placeholder="John"
                              value={formData.first_name}
                              onChange={handleInputChange}
                              className={`pl-12 h-12 text-base border-slate-300 dark:border-slate-600 focus:border-purple-500 dark:focus:border-purple-400 focus:ring-2 focus:ring-purple-500/20 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 transition-all duration-200 ${
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
                          <Label htmlFor="last_name" className="text-base font-semibold text-slate-700 dark:text-slate-300">
                            Last Name *
                          </Label>
                          <div className="relative group">
                            <User className="absolute left-4 top-4 h-5 w-5 text-slate-400 group-focus-within:text-purple-500 transition-colors" />
                            <Input
                              id="last_name"
                              name="last_name"
                              type="text"
                              placeholder="Doe"
                              value={formData.last_name}
                              onChange={handleInputChange}
                              className={`pl-12 h-12 text-base border-slate-300 dark:border-slate-600 focus:border-purple-500 dark:focus:border-purple-400 focus:ring-2 focus:ring-purple-500/20 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 transition-all duration-200 ${
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
                      <div className="text-center mb-8">
                        <h3 className="text-2xl font-bold bg-gradient-to-r from-purple-600 to-blue-600 bg-clip-text text-transparent">
                          Account Setup
                        </h3>
                        <p className="text-sm text-slate-600 dark:text-slate-400 mt-2">Set up your login credentials</p>
                      </div>

                      <div className="space-y-5">
                        <div className="space-y-2">
                          <Label htmlFor="email" className="text-base font-semibold text-slate-700 dark:text-slate-300">
                            Email Address * {invitationToken && <span className="text-sm text-purple-600">(From Invitation)</span>}
                          </Label>
                          <div className="relative group">
                            <Mail className="absolute left-4 top-4 h-5 w-5 text-slate-400 group-focus-within:text-purple-500 transition-colors" />
                            <Input
                              id="email"
                              name="email"
                              type="email"
                              placeholder="john.doe@company.com"
                              value={formData.email}
                              onChange={handleInputChange}
                              className={`pl-12 h-12 text-base border-slate-300 dark:border-slate-600 focus:border-purple-500 dark:focus:border-purple-400 focus:ring-2 focus:ring-purple-500/20 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 transition-all duration-200 ${
                                validationErrors.email ? 'border-red-500 focus:border-red-500' : ''
                              } ${invitationToken ? 'bg-slate-100 dark:bg-slate-700 cursor-not-allowed' : ''}`}
                              disabled={loading}
                              readOnly={!!invitationToken}
                              required
                            />
                          </div>
                          {validationErrors.email && (
                            <p className="text-xs text-red-600 flex items-center gap-1">
                              <AlertCircle className="h-3 w-3" />
                              {validationErrors.email}
                            </p>
                          )}
                        </div>

                        <div className="space-y-2">
                          <Label htmlFor="phone" className="text-base font-semibold text-slate-700 dark:text-slate-300">
                            Phone Number
                          </Label>
                          <div className="relative group">
                            <Phone className="absolute left-4 top-4 h-5 w-5 text-slate-400 group-focus-within:text-purple-500 transition-colors" />
                            <Input
                              id="phone"
                              name="phone"
                              type="tel"
                              placeholder="+1 (555) 123-4567"
                              value={formData.phone}
                              onChange={handleInputChange}
                              className="pl-12 h-12 text-base border-slate-300 dark:border-slate-600 focus:border-purple-500 dark:focus:border-purple-400 focus:ring-2 focus:ring-purple-500/20 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 transition-all duration-200"
                              disabled={loading}
                            />
                          </div>
                        </div>

                        <div className="mt-6 p-4 rounded-xl bg-gradient-to-br from-purple-50 to-blue-50 dark:from-purple-900/20 dark:to-blue-900/20 border border-purple-200 dark:border-purple-800">
                          <p className="text-sm text-slate-700 dark:text-slate-300 flex items-center gap-2">
                            <Users className="h-4 w-4 text-purple-600" />
                            <span className="font-medium">After registration, you'll create or join an organization</span>
                          </p>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* Step 3: Security */}
                  {currentStep === 3 && (
                    <div className="space-y-6">
                      <div className="text-center mb-8">
                        <h3 className="text-2xl font-bold bg-gradient-to-r from-purple-600 to-blue-600 bg-clip-text text-transparent">
                          Security
                        </h3>
                        <p className="text-sm text-slate-600 dark:text-slate-400 mt-2">Create a strong password to protect your account</p>
                      </div>

                      <div className="space-y-5">
                        <div className="space-y-2">
                          <Label htmlFor="password" className="text-base font-semibold text-slate-700 dark:text-slate-300">
                            Password *
                          </Label>
                          <div className="relative group">
                            <Lock className="absolute left-4 top-4 h-5 w-5 text-slate-400 group-focus-within:text-purple-500 transition-colors" />
                            <Input
                              id="password"
                              name="password"
                              type={showPassword ? 'text' : 'password'}
                              placeholder="Create a strong password"
                              value={formData.password}
                              onChange={handleInputChange}
                              className={`pl-12 pr-12 h-12 text-base border-slate-300 dark:border-slate-600 focus:border-purple-500 dark:focus:border-purple-400 focus:ring-2 focus:ring-purple-500/20 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 transition-all duration-200 ${
                                validationErrors.password ? 'border-red-500 focus:border-red-500' : ''
                              }`}
                              disabled={loading}
                              required
                            />
                            <button
                              type="button"
                              onClick={togglePasswordVisibility}
                              className="absolute right-4 top-3.5 text-slate-400 hover:text-purple-600 transition-colors"
                              disabled={loading}
                            >
                              {showPassword ? (
                                <EyeOff className="h-5 w-5" />
                              ) : (
                                <Eye className="h-5 w-5" />
                              )}
                            </button>
                          </div>

                          {/* Password Strength Indicator */}
                          {formData.password && (
                            <div className="space-y-3 mt-4">
                              <div className="flex items-center justify-between">
                                <span className="text-sm font-semibold text-slate-700 dark:text-slate-200">Password Strength</span>
                                <Badge variant="outline" className={`text-xs font-semibold border-2 px-3 py-1 ${
                                  passwordStrengthScore <= 2 ? 'border-red-400 text-red-700 dark:border-red-500 dark:text-red-300 bg-red-50 dark:bg-red-950' :
                                  passwordStrengthScore <= 3 ? 'border-yellow-400 text-yellow-700 dark:border-yellow-500 dark:text-yellow-300 bg-yellow-50 dark:bg-yellow-950' :
                                  passwordStrengthScore <= 4 ? 'border-blue-400 text-blue-700 dark:border-blue-500 dark:text-blue-300 bg-blue-50 dark:bg-blue-950' :
                                  'border-emerald-400 text-emerald-700 dark:border-emerald-500 dark:text-emerald-300 bg-emerald-50 dark:bg-emerald-950'
                                }`}>
                                  {getPasswordStrengthLabel()}
                                </Badge>
                              </div>
                              <div className="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-3 overflow-hidden">
                                <motion.div
                                  className={`h-3 rounded-full ${getPasswordStrengthColor()}`}
                                  initial={{ width: 0 }}
                                  animate={{ width: `${passwordStrengthPercentage}%` }}
                                  transition={{ duration: 0.3 }}
                                />
                              </div>
                              <div className="grid grid-cols-1 gap-2 text-sm mt-3">
                                {[
                                  { key: 'length', label: 'At least 8 characters' },
                                  { key: 'uppercase', label: 'One uppercase letter' },
                                  { key: 'lowercase', label: 'One lowercase letter' },
                                  { key: 'number', label: 'One number' },
                                  { key: 'special', label: 'One special character' }
                                ].map(({ key, label }) => (
                                  <motion.div 
                                    key={key} 
                                    className={`flex items-center gap-2 transition-all duration-200 ${
                                      passwordStrength[key as keyof typeof passwordStrength] ? 'text-emerald-700 dark:text-emerald-300' : 'text-slate-500 dark:text-slate-400'
                                    }`}
                                    initial={{ opacity: 0, x: -10 }}
                                    animate={{ opacity: 1, x: 0 }}
                                    transition={{ delay: 0.1 }}
                                  >
                                    {passwordStrength[key as keyof typeof passwordStrength] ? (
                                      <CheckCircle className="h-4 w-4" />
                                    ) : (
                                      <div className="h-4 w-4 rounded-full border-2 border-current" />
                                    )}
                                    <span className="font-medium">{label}</span>
                                  </motion.div>
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
                          <Label htmlFor="confirm_password" className="text-base font-semibold text-slate-700 dark:text-slate-300">
                            Confirm Password *
                          </Label>
                          <div className="relative group">
                            <Lock className="absolute left-4 top-4 h-5 w-5 text-slate-400 group-focus-within:text-purple-500 transition-colors" />
                            <Input
                              id="confirm_password"
                              name="confirm_password"
                              type={showConfirmPassword ? 'text' : 'password'}
                              placeholder="Confirm your password"
                              value={formData.confirm_password}
                              onChange={handleInputChange}
                              className={`pl-12 pr-12 h-12 text-base border-slate-300 dark:border-slate-600 focus:border-purple-500 dark:focus:border-purple-400 focus:ring-2 focus:ring-purple-500/20 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 transition-all duration-200 ${
                                validationErrors.confirm_password ? 'border-red-500 focus:border-red-500' : ''
                              }`}
                              disabled={loading}
                              required
                            />
                            <button
                              type="button"
                              onClick={toggleConfirmPasswordVisibility}
                              className="absolute right-4 top-3.5 text-slate-400 hover:text-purple-600 transition-colors"
                              disabled={loading}
                            >
                              {showConfirmPassword ? (
                                <EyeOff className="h-5 w-5" />
                              ) : (
                                <Eye className="h-5 w-5" />
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
              <div className="flex justify-between pt-8 border-t border-slate-200 dark:border-slate-700">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setCurrentStep(Math.max(1, currentStep - 1))}
                  disabled={currentStep === 1 || loading}
                  className="px-8 h-12 text-base font-semibold border-slate-300 hover:bg-slate-50 dark:border-slate-600 dark:hover:bg-slate-800"
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
                      // Email validation with + character support
                      const emailRegex = /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/
                      if (currentStep === 2 && (!formData.email.trim() || !emailRegex.test(formData.email))) {
                        setValidationErrors({
                          email: !formData.email.trim() ? 'Email is required' :
                                 !emailRegex.test(formData.email) ? 'Please enter a valid email' : ''
                        })
                        return
                      }
                      setCurrentStep(currentStep + 1)
                    }}
                    disabled={loading}
                    className="px-8 h-12 text-base font-semibold bg-gradient-to-r from-purple-600 via-blue-600 to-indigo-600 hover:from-purple-700 hover:via-blue-700 hover:to-indigo-700 text-white shadow-lg hover:shadow-xl transition-all duration-300"
                  >
                    Next
                    <ArrowRight className="w-5 h-5 ml-2 group-hover:translate-x-1 transition-transform" />
                  </Button>
                ) : (
                  <Button
                    type="submit"
                    disabled={loading || !isPasswordValid}
                    className="px-8 h-12 text-base bg-gradient-to-r from-purple-600 via-blue-600 to-indigo-600 hover:from-purple-700 hover:via-blue-700 hover:to-indigo-700 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed group"
                  >
                    {loading ? (
                      <>
                        <LoadingSpinner size="sm" className="mr-2" />
                        Creating Account...
                      </>
                    ) : (
                      <>
                        Create Account
                        <Check className="w-5 h-5 ml-2 group-hover:scale-110 transition-transform" />
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
              className="mt-8 text-center pb-4"
            >
              <span className="text-slate-600 dark:text-slate-400 text-base">Already have an account? </span>
              <Link
                to="/login"
                className="text-purple-600 dark:text-purple-400 hover:text-purple-700 dark:hover:text-purple-300 font-semibold transition-colors inline-flex items-center gap-1 group"
              >
                Sign in here
                <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
              </Link>
            </motion.div>
          </CardContent>
        </Card>

        {/* Footer */}
        <motion.p
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.6 }}
          className="text-center text-slate-600 dark:text-slate-400 mt-8 text-sm"
        >
          Protected by enterprise-grade security
        </motion.p>
      </motion.div>
    </div>
  )
}

export default Register
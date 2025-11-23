import { NextRequest, NextResponse } from 'next/server';

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { email, password, confirm_password, first_name, last_name, phone, address } = body;

    // Validate required fields
    if (!email || !password || !first_name || !last_name) {
      return NextResponse.json(
        { error: 'Missing required fields', message: 'Email, password, first name, and last name are required' },
        { status: 400 }
      );
    }

    // Validate email format
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      return NextResponse.json(
        { error: 'Invalid email', message: 'Please provide a valid email address' },
        { status: 400 }
      );
    }

    // Validate password strength
    if (password.length < 8) {
      return NextResponse.json(
        { error: 'Weak password', message: 'Password must be at least 8 characters long' },
        { status: 400 }
      );
    }

    // Forward the registration request to the Go backend
    const backendUrl = process.env.NODE_ENV === 'development' 
      ? 'http://localhost:8080' 
      : process.env.BACKEND_URL || '';

    console.log('Forwarding registration to backend:', `${backendUrl}/api/v1/auth/register`);

    const backendResponse = await fetch(`${backendUrl}/api/v1/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        email,
        password,
        confirm_password,
        first_name,
        last_name,
        phone: phone || null,
        address: address || null,
      }),
    });

    const backendData = await backendResponse.json();

    if (!backendResponse.ok) {
      console.error('Backend registration error:', backendData);
      
      // Handle specific backend errors
      if (backendResponse.status === 409) {
        return NextResponse.json(
          { error: 'User exists', message: 'An account with this email already exists' },
          { status: 409 }
        );
      }
      
      return NextResponse.json(
        { 
          error: 'Registration failed', 
          message: backendData.message || 'Failed to create account. Please try again.' 
        },
        { status: backendResponse.status }
      );
    }

    console.log('Registration successful:', { email, user_id: backendData.data?.user?.id });

    // Return success response
    return NextResponse.json({
      success: true,
      message: 'Account created successfully',
      data: {
        user: backendData.data?.user || {
          email,
          first_name,
          last_name,
          phone,
          address
        }
      }
    }, { status: 201 });

  } catch (error) {
    console.error('Registration API error:', error);
    
    // Handle network or parsing errors
    if (error instanceof TypeError && error.message.includes('fetch')) {
      return NextResponse.json(
        { 
          error: 'Service unavailable', 
          message: 'Unable to connect to authentication service. Please try again later.' 
        },
        { status: 503 }
      );
    }

    return NextResponse.json(
      { 
        error: 'Internal error', 
        message: 'An unexpected error occurred. Please try again.' 
      },
      { status: 500 }
    );
  }
}
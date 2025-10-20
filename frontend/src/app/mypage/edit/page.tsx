'use client';

import React from 'react';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { ProfileEditContent } from '@/components/mypage';

const ProfileEditPage: React.FC = () => {
  return (
    <ProtectedRoute>
      <ProfileEditContent />
    </ProtectedRoute>
  );
};

export default ProfileEditPage;

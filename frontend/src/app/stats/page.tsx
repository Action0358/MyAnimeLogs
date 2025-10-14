'use client';

import React from 'react';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { StatsPageContent } from '@/components/stats';

const StatsPage: React.FC = () => {
  return (
    <ProtectedRoute>
      <StatsPageContent />
    </ProtectedRoute>
  );
};

export default StatsPage;
